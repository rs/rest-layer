package query

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type predicateParser struct {
	query string
	pos   int
}

// MustParsePredicate parses a predicate expression and panics in case of error.
func MustParsePredicate(query string) Predicate {
	q, err := ParsePredicate(query)
	if err != nil {
		panic(fmt.Sprintf("query: ParsePredicate(%q): %v", query, err))
	}
	return q
}

// ParsePredicate parses a predicate.
func ParsePredicate(predicate string) (Predicate, error) {
	if predicate == "" {
		return Predicate{}, nil
	}
	p := &predicateParser{query: predicate}
	return p.parse()
}

func (p *predicateParser) parse() (Predicate, error) {
	p.eatWhitespaces()
	q, err := p.parseExpressions()
	if err != nil {
		return nil, fmt.Errorf("char %d: %v", p.pos, err)
	}
	p.eatWhitespaces()
	if p.more() {
		return nil, fmt.Errorf("char %d: expected EOF got %q", p.pos, p.peek())
	}
	return q, nil
}

// parseExpressions parses one or more expression enclosed inside brackets.
//
// Examples:
//   {foo: "bar"}
//   {foo: "bar", bar: "baz"}
//   {$or: [{foo: "bar"}, {foo: "baz"}]}
//   {foo: {$exists: true}}
func (p *predicateParser) parseExpressions() ([]Expression, error) {
	exps := []Expression{}
	if !p.expect('{') {
		return nil, fmt.Errorf("expected '{' got %q", p.peek())
	}
	p.eatWhitespaces()
	if p.expect('}') {
		// Empty expression block
		return exps, nil
	}
	for {
		p.eatWhitespaces()
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		exps = append(exps, exp)
		p.eatWhitespaces()
		if !p.expect(',') {
			break
		}
	}
	if !p.expect('}') {
		return nil, fmt.Errorf("expected '}' got %q", p.peek())
	}
	return exps, nil
}

// parseExpression parses an expression without surrounding braces.
//
// Examples of expressions:
//   foo: "bar"
//   $or: [{foo: "bar"}, {foo: "baz"}]
//   foo: {$exists: true}
func (p *predicateParser) parseExpression() (Expression, error) {
	oldPos := p.pos
	label, err := p.parseLabel()
	if err != nil {
		return nil, err
	}
	p.eatWhitespaces()
	switch label {
	case opAnd, opOr:
		subExp, err := p.parseSubExpressions()
		if err != nil {
			return nil, fmt.Errorf("%s: %v", label, err)
		}
		if len(subExp) < 1 {
			return nil, fmt.Errorf("%s: one expressions or more required", label)
		}
		if label == opAnd {
			and := And(subExp)
			return &and, nil
		}
		or := Or(subExp)
		return &or, nil
	case opExists, opIn, opNotIn, opNotEqual, opRegex, opElemMatch,
		opLowerThan, opLowerOrEqual, opGreaterThan, opGreaterOrEqual:
		p.pos = oldPos
		return nil, fmt.Errorf("%s: invalid placement", label)
	default:
		exp, err := p.parseCommand(label)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", label, err)
		}
		return exp, nil
	}
}

// parseSubExpressions parses [{exp}, {exp, exp}...].
func (p *predicateParser) parseSubExpressions() ([]Expression, error) {
	if !p.expect('[') {
		return nil, fmt.Errorf("expected '[' got %q", p.peek())
	}
	subExps := []Expression{}
	p.eatWhitespaces()
	if p.expect(']') {
		// Empty list
		return subExps, nil
	}
	for {
		p.eatWhitespaces()
		exps, err := p.parseExpressions()
		if err != nil {
			return nil, err
		}
		switch len(exps) {
		case 0:
		// XXX empty expression?
		case 1:
			subExps = append(subExps, exps[0])
		default:
			and := And(exps)
			subExps = append(subExps, &and)
		}
		p.eatWhitespaces()
		if !p.expect(',') {
			break
		}
	}
	if !p.expect(']') {
		return nil, fmt.Errorf("expected ']' got %q", p.peek())
	}
	return subExps, nil
}

// parseCommand parses the command that come after a label.
//
// Examples of commands:
//   "foo" // equal
//   {"foo": "bar"} // equal
//   ["foo", "bar"] // equal
//   {$exist: true}
//   {$ne: "foo"}
//   {$in: ["foo", "bar"]}
func (p *predicateParser) parseCommand(field string) (Expression, error) {
	oldPos := p.pos
	if p.expect('{') {
		p.eatWhitespaces()
		if p.expect('}') {
			// Empty dict must be parsed as a value
			goto VALUE
		}
		label, err := p.parseLabel()
		if err != nil {
			return nil, err
		}
		p.eatWhitespaces()
		switch label {
		case opExists:
			v, err := p.parseBool()
			if err != nil {
				return nil, fmt.Errorf("%s: %v", label, err)
			}
			p.eatWhitespaces()
			if !p.expect('}') {
				return nil, fmt.Errorf("%s: expected '}' got %q", label, p.peek())
			}
			if v {
				return &Exist{Field: field}, nil
			}
			return &NotExist{Field: field}, nil
		case opIn, opNotIn:
			values, err := p.parseValues()
			if err != nil {
				return nil, fmt.Errorf("%s: %v", label, err)
			}
			p.eatWhitespaces()
			if !p.expect('}') {
				return nil, fmt.Errorf("%s: expected '}' got %q", label, p.peek())
			}
			if label == opIn {
				return &In{Field: field, Values: values}, nil
			}
			return &NotIn{Field: field, Values: values}, nil
		case opNotEqual:
			value, err := p.parseValue()
			if err != nil {
				return nil, fmt.Errorf("%s: %v", label, err)
			}
			p.eatWhitespaces()
			if !p.expect('}') {
				return nil, fmt.Errorf("%s: expected '}' got %q", label, p.peek())
			}
			return &NotEqual{Field: field, Value: value}, nil
		case opLowerThan, opLowerOrEqual, opGreaterThan, opGreaterOrEqual:
			value, err := p.parseValue()
			if err != nil {
				return nil, fmt.Errorf("%s: %v", label, err)
			}
			p.eatWhitespaces()
			if !p.expect('}') {
				return nil, fmt.Errorf("%s: expected '}' got %q", label, p.peek())
			}
			switch label {
			case opLowerThan:
				return &LowerThan{Field: field, Value: value}, nil
			case opLowerOrEqual:
				return &LowerOrEqual{Field: field, Value: value}, nil
			case opGreaterThan:
				return &GreaterThan{Field: field, Value: value}, nil
			case opGreaterOrEqual:
				return &GreaterOrEqual{Field: field, Value: value}, nil
			}
		case opRegex:
			str, err := p.parseString()
			if err != nil {
				return nil, fmt.Errorf("%s: %v", label, err)
			}
			re, err := regexp.Compile(str)
			if err != nil {
				return nil, fmt.Errorf("%s: invalid regex: %v", label, err)
			}
			p.eatWhitespaces()
			if !p.expect('}') {
				return nil, fmt.Errorf("%s: expected '}' got %q", label, p.peek())
			}
			return &Regex{Field: field, Value: re}, nil
		case opElemMatch:
			exps, err := p.parseExpressions()
			if err != nil {
				return nil, fmt.Errorf("%s: %v", label, err)
			}
			p.eatWhitespaces()
			if !p.expect('}') {
				return nil, fmt.Errorf("%s: expected '}' got %q", label, p.peek())
			}
			return &ElemMatch{Field: field, Exps: exps}, nil
		}
	}
VALUE:
	// If the current position is not a dictionary ({}) or is a dictionary with
	// no known command, parse the next chars as value for an equal command.
	p.pos = oldPos // restore cursor to initial state
	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	return &Equal{Field: field, Value: value}, nil
}

// parseLabel parses a label with or without quotes and advance the curser right
// after the ":".
func (p *predicateParser) parseLabel() (label string, err error) {
	if p.peek() == '"' {
		if label, err = p.parseString(); err != nil {
			return "", fmt.Errorf("invalid label: %v", err)
		}
	} else {
		// Try to parse unquoted label
		end := p.pos
		for ; end < len(p.query); end++ {
			if c := p.peekAt(end); (c >= '0' && c <= '9') ||
				(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
				c == '$' || c == '.' || c == '_' || c == '-' {
				continue
			}
			break
		}
		if p.pos == end {
			return "", fmt.Errorf("expected a label got %q", p.peek())
		}
		label = p.query[p.pos:end]
		p.pos = end
	}
	p.eatWhitespaces()
	if !p.expect(':') {
		return "", fmt.Errorf("expected ':' got %q", p.peek())
	}
	return label, nil
}

// parseValue parses a value like JSON with optional quotes for dictionary keys
// and advance the cursor after the end of the value.
func (p *predicateParser) parseValue() (Value, error) {
	c := p.peek()
	switch c {
	case '"':
		return p.parseString()
	case '{':
		return p.parseDict()
	case '[':
		return p.parseValues()
	case 't', 'f':
		return p.parseBool()
	case 'n':
		return p.parseNull()
	default:
		if (c >= '0' && c <= '9') || c == '-' {
			// Parse a number
			return p.parseNumber()
		}
	}
	return nil, fmt.Errorf("unexpected char %q", c)
}

// parseValues parses a list of values like [Value, Value...].
func (p *predicateParser) parseValues() ([]Value, error) {
	if !p.expect('[') {
		return nil, fmt.Errorf("expected '[' got %q", p.peek())
	}
	values := []Value{}
	p.eatWhitespaces()
	if p.expect(']') {
		// Empty
		return values, nil
	}
	for {
		p.eatWhitespaces()
		value, err := p.parseValue()
		if err != nil {
			return nil, fmt.Errorf("item #%d: %v", len(values), err)
		}
		values = append(values, value)
		p.eatWhitespaces()
		if !p.expect(',') {
			break
		}
	}
	if !p.expect(']') {
		return nil, fmt.Errorf("expected ',' or ']' got %q", p.peek())
	}
	return values, nil
}

// parseDict parses a dictionary of key: Value with keys optionally quotted.
func (p *predicateParser) parseDict() (map[string]Value, error) {
	if !p.expect('{') {
		return nil, fmt.Errorf("expected '{' got %q", p.peek())
	}
	dict := map[string]Value{}
	p.eatWhitespaces()
	if p.expect('}') {
		// Empty dict
		return dict, nil
	}
	for {
		p.eatWhitespaces()
		label, err := p.parseLabel()
		if err != nil {
			return nil, err
		}
		p.eatWhitespaces()
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		dict[label] = value
		p.eatWhitespaces()
		if !p.expect(',') {
			break
		}
	}
	if !p.expect('}') {
		return nil, fmt.Errorf("expected '}' got %q", p.peek())
	}
	return dict, nil
}

// parseBool parses a Boolean value.
func (p *predicateParser) parseBool() (bool, error) {
	switch p.peek() {
	case 't':
		if p.pos+4 <= len(p.query) && p.query[p.pos:p.pos+4] == "true" {
			p.pos += 4
			return true, nil
		}
	case 'f':
		if p.pos+5 <= len(p.query) && p.query[p.pos:p.pos+5] == "false" {
			p.pos += 5
			return false, nil
		}
	}
	return false, errors.New("not a boolean")
}

// parseNull parses a JSON Null value.
func (p *predicateParser) parseNull() (Value, error) {
	c := p.peek()
	if c == 'n' && p.pos+4 <= len(p.query) && p.query[p.pos:p.pos+4] == "null" {
		p.pos += 4
		return nil, nil
	}
	return nil, errors.New("not null")
}

// parseNumber parses a number as float.
func (p *predicateParser) parseNumber() (float64, error) {
	end := p.pos
	if p.peek() == '-' {
		end++
	}
	if c := p.peekAt(end); c >= '0' && c <= '9' {
		end++
	} else {
		return 0, errors.New("not a number")
	}
	for {
		if c := p.peekAt(end); (c >= '0' && c <= '9') || c == '.' || c == 'e' {
			end++
		} else {
			break
		}
	}
	f, err := strconv.ParseFloat(p.query[p.pos:end], 64)
	if err != nil {
		return 0, fmt.Errorf("not a number: %v", err)
	}
	p.pos = end
	return f, nil
}

// parseString parses a string.
func (p *predicateParser) parseString() (string, error) {
	if p.peek() != '"' {
		return "", errors.New("not a string")
	}
	start := p.pos
	end := start + 1
	quoted := false
	simple := true
	done := false
	for ; end < len(p.query); end++ {
		c := p.peekAt(end)
		if quoted {
			quoted = false
		} else if c == '\\' {
			quoted = true
			simple = false
		} else if c == '"' {
			done = true
			end++
			break
		}
	}
	if !done {
		return "", errors.New("not a string: unexpected EOF")
	}
	if simple {
		// If no quoted char found, just return the string without the surrounding
		// quotes to avoid allocation.
		p.pos = end
		return p.query[start+1 : end-1], nil
	}
	s, err := strconv.Unquote(p.query[start:end])
	if err != nil {
		return "", fmt.Errorf("not a string: %v", err)
	}
	p.pos = end
	return s, nil
}

// more returns true if there is more data to parse.
func (p *predicateParser) more() bool {
	return p.pos < len(p.query)
}

// expect advances the cursor if the current char is equal to c or return
// false otherwise.
func (p *predicateParser) expect(c byte) bool {
	if p.peek() == c {
		p.pos++
		return true
	}
	return false
}

// peek returns the char at the current position without advancing the cursor.
func (p *predicateParser) peek() byte {
	if p.more() {
		return p.query[p.pos]
	}
	return 0
}

// peek returns the char at the given position without moving the cursor.
func (p *predicateParser) peekAt(pos int) byte {
	if pos < len(p.query) {
		return p.query[pos]
	}
	return 0
}

// eatWhitespaces advance the cursor position pos until non printable characters are met.
func (p *predicateParser) eatWhitespaces() {
	for p.more() {
		switch p.query[p.pos] {
		case ' ', '\n', '\r', '\t':
			p.pos++
			continue
		}
		break
	}
}
