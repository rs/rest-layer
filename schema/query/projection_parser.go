package query

import (
	"errors"
	"fmt"
	"strconv"
)

type projectionParser struct {
	exp string
	pos int
}

/*
ParseProjection recursively parses a projection expression.

Projection expression syntax allows to list fields that must be kept in the
response hierarchically.

A field is an alphanum + - and _ separated by comas:

field1,field2

When a document has sub-fields, sub-resources or sub-connections, the
sub-element's fields can be specified as well by enclosing them between braces:

field1{sub-field1,sub-field2},field2

Fields can get some some parameters which can be passed to field filters to
transform the value. Parameters are passed as key:value pairs enclosed in
parenthesizes, with value being either a quotted string or a numerical value:

field1(param1:"value", param2:123),field2

You can combine field params and sub-field definition:

field1(param1:"value", param2:123){sub-field1,sub-field2},field2

Or pass params to sub-fields:

field1{sub-field1(param1:"value"),sub-field2},field2

Fields can also be renamed (aliased). This is useful when you want to have
several times the same fields with different sets of parameters. To define
aliases, prepend the field definition by the alias name and a colon (:):

field:alias

With params:

thumbnail_small_url:thumbnail_url(size=80),thumbnail_small_url:thumbnail_url(size=500)

With this example, the resulted document would be:

  {
    "thumbnail_small_url": "the url with size 80",
    "thumbnail_large_url": "the url with size 500",
  }

*/
func ParseProjection(projection string) (Projection, error) {
	p := &projectionParser{exp: projection}
	return p.parse()
}

// MustParseProjection parses a projection expression and panics in case of error.
func MustParseProjection(projection string) Projection {
	p, err := ParseProjection(projection)
	if err != nil {
		panic(fmt.Sprintf("query: ParseProjection(%q): %v", projection, err))
	}
	return p
}

func (p *projectionParser) parse() (Projection, error) {
	return p.parseExpression(false)
}

func (p *projectionParser) parseExpression(opened bool) (Projection, error) {
	expectField := false
	projection := []ProjectionField{}
	var field *ProjectionField
	for p.more() {
		if field == nil {
			p.eatWhitespaces()
			name, alias := p.scanFieldNameWithAlias()
			if name == "" {
				return nil, fmt.Errorf("looking for field name at char %d", p.pos)
			}
			field = &ProjectionField{Name: name, Alias: alias}
			expectField = false
			continue
		}
		switch p.peek() {
		case '{':
			p.pos++
			children, err := p.parseExpression(true)
			if err != nil {
				return nil, err
			}
			field.Children = children
		case '}':
			if opened && !expectField {
				projection = append(projection, *field)
				return projection, nil
			}
			return nil, fmt.Errorf("looking for field name and got `}' at char %d", p.pos)
		case '(':
			p.pos++
			params, err := p.scanFieldParams()
			if err != nil {
				return nil, err
			}
			field.Params = params
		case ',':
			projection = append(projection, *field)
			field = nil
			expectField = true
		case ' ', '\n', '\r', '\t':
			// ignore whitespace
		default:
			return nil, fmt.Errorf("invalid char `%c` at %d", p.peek(), p.pos)
		}
		p.pos++
	}
	if expectField {
		return nil, fmt.Errorf("looking for field name at char %d", p.pos)
	}
	if opened {
		return nil, fmt.Errorf("looking for `}' at char %d", p.pos)
	}
	if field != nil {
		projection = append(projection, *field)
	}
	return projection, nil
}

// p.scanFieldParams parses fields params until it finds a closing
// parenthesis. If the max length is reached before or a syntax error is found,
// an error is returned.
//
// It gets the expression buffer as "exp", the current position after an opening
// parenthesis at pos.
func (p *projectionParser) scanFieldParams() (map[string]interface{}, error) {
	params := map[string]interface{}{}
	for p.more() {
		p.eatWhitespaces()
		name := p.scanFieldName()
		if name == "" {
			return nil, fmt.Errorf("looking for parameter name at char %d", p.pos)
		}
		p.eatWhitespaces()
		if !p.expect(':') && !p.expect('=') {
			return nil, fmt.Errorf("looking for : at char %d", p.pos)
		}
		p.eatWhitespaces()
		value, err := p.scanParamValue()
		if err != nil {
			return nil, err
		}
		params[name] = value
		p.eatWhitespaces()
		c := p.peek()
		if c == ')' {
			break
		} else if c == ',' {
			p.pos++
		} else {
			return nil, fmt.Errorf("looking for `,' or ')' at char %d", p.pos)
		}
	}
	return params, nil
}

// p.scanFieldName captures a field name at current position and advance
// the cursor position "pos" at the next character following the field name.
func (p *projectionParser) scanFieldName() string {
	field := []byte{}
	for p.more() {
		c := p.peek()
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '*' {
			// Allow * only by itself.
			if (c == '*' && len(field) != 0) || (c != '*' && len(field) > 0 && field[0] == '*') {
				return ""
			}
			field = append(field, c)
			p.pos++
			continue
		}
		break
	}
	return string(field)
}

// p.scanFieldNameWithAlias parses a field optional alias followed by it's
// name separated by a column at current position and advance the cursor
// position "pos" at the next character following the field name.
func (p *projectionParser) scanFieldNameWithAlias() (name string, alias string) {
	name = p.scanFieldName()
	p.eatWhitespaces()
	if p.expect(':') {
		p.eatWhitespaces()
		alias = name
		name = p.scanFieldName()
	}
	return name, alias
}

// p.scanParamValue captures a parameter value at the current position and
// advance the cursor position "pos" at the next character following the field name.
//
// The returned value may be either a string if the value was quotted or a float
// if not an was a valid number. In case of syntax error, an error is returned.
func (p *projectionParser) scanParamValue() (interface{}, error) {
	switch p.peek() {
	case '"', '\'':
		return p.parseString()
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		return p.parseNumber()
	case 't', 'f':
		return p.parseBool()
	default:
		return nil, fmt.Errorf("looking for value at char %d", p.pos)
	}
}

// parseBool parses a Boolean value.
func (p *projectionParser) parseBool() (bool, error) {
	switch p.peek() {
	case 't':
		if p.pos+4 <= len(p.exp) && p.exp[p.pos:p.pos+4] == "true" {
			p.pos += 4
			return true, nil
		}
	case 'f':
		if p.pos+5 <= len(p.exp) && p.exp[p.pos:p.pos+5] == "false" {
			p.pos += 5
			return false, nil
		}
	}
	return false, errors.New("not a boolean")
}

// parseNumber parses a number as float.
func (p *projectionParser) parseNumber() (float64, error) {
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
	f, err := strconv.ParseFloat(p.exp[p.pos:end], 64)
	if err != nil {
		return 0, fmt.Errorf("not a number: %v", err)
	}
	p.pos = end
	return f, nil
}

// parseString parses a string quotted with " or '.
func (p *projectionParser) parseString() (string, error) {
	quote := p.peek()
	if quote != '"' && quote != '\'' {
		return "", errors.New("not a string")
	}
	start := p.pos
	end := start + 1
	escaped := false
	simple := true
	done := false
	for ; end < len(p.exp); end++ {
		c := p.peekAt(end)
		if escaped {
			escaped = false
		} else if c == '\\' {
			escaped = true
			simple = false
		} else if c == quote {
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
		return p.exp[start+1 : end-1], nil
	}
	s, err := strconv.Unquote(p.exp[start:end])
	if err != nil {
		return "", fmt.Errorf("not a string: %v", err)
	}
	p.pos = end
	return s, nil
}

// more returns true if there is more data to parse.
func (p *projectionParser) more() bool {
	return p.pos < len(p.exp)
}

// expect advances the cursor if the current char is equal to c or return
// false otherwise.
func (p *projectionParser) expect(c byte) bool {
	if p.peek() == c {
		p.pos++
		return true
	}
	return false
}

// peek returns the char at the current position without advancing the cursor.
func (p *projectionParser) peek() byte {
	if p.more() {
		return p.exp[p.pos]
	}
	return 0
}

// peek returns the char at the given position without moving the cursor.
func (p *projectionParser) peekAt(pos int) byte {
	if pos < len(p.exp) {
		return p.exp[pos]
	}
	return 0
}

// eatWhitespaces advance the cursor position pos until non printable characters are met.
func (p *projectionParser) eatWhitespaces() {
	for p.more() {
		switch p.exp[p.pos] {
		case ' ', '\n', '\r', '\t':
			p.pos++
			continue
		}
		break
	}
}
