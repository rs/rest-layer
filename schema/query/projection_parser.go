package query

import (
	"fmt"
	"strconv"
)

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
	pos := 0
	return parseProjectionExpression(projection, &pos, false)
}

func parseProjectionExpression(exp string, pos *int, opened bool) (Projection, error) {
	expectField := false
	projection := []ProjectionField{}
	var field *ProjectionField
	for *pos < len(exp) {
		if field == nil {
			name, alias := scanProjectionFieldNameWithAlias(exp, pos)
			if name == "" {
				return nil, fmt.Errorf("looking for field name at char %d", *pos)
			}
			field = &ProjectionField{Name: name, Alias: alias}
			expectField = false
			continue
		}
		c := exp[*pos]
		switch c {
		case '{':
			*pos++
			children, err := parseProjectionExpression(exp, pos, true)
			if err != nil {
				return nil, err
			}
			field.Children = children
		case '}':
			if opened && !expectField {
				projection = append(projection, *field)
				return projection, nil
			}
			return nil, fmt.Errorf("looking for field name and got `}' at char %d", *pos)
		case '(':
			*pos++
			params, err := scanProjectionFieldParams(exp, pos)
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
			return nil, fmt.Errorf("invalid char `%c` at %d", c, *pos)
		}
		*pos++
	}
	if expectField {
		return nil, fmt.Errorf("looking for field name at char %d", *pos)
	}
	if opened {
		return nil, fmt.Errorf("looking for `}' at char %d", *pos)
	}
	if field != nil {
		projection = append(projection, *field)
	}
	return projection, nil
}

// scanProjectionFieldParams parses fields params until it finds a closing
// parenthesis. If the max length is reached before or a syntax error is found,
// an error is returned.
//
// It gets the expression buffer as "exp", the current position after an opening
// parenthesis at pos.
func scanProjectionFieldParams(exp string, pos *int) (map[string]interface{}, error) {
	params := map[string]interface{}{}
	for *pos < len(exp) {
		name := scanProjectionFieldName(exp, pos)
		if name == "" {
			return nil, fmt.Errorf("looking for parameter name at char %d", *pos)
		}
		found := false
	L:
		for *pos < len(exp) {
			c := exp[*pos]
			switch c {
			case ':', '=': // accept both : and = for backward compat
				found = true
				break L
			case ' ', '\n', '\r', '\t':
				// ignore whitespaces
			default:
				return nil, fmt.Errorf("looking for : at char %d", *pos)
			}
			*pos++
		}
		if !found {
			return nil, fmt.Errorf("looking for : at char %d", *pos)
		}
		*pos++
		value, err := scanProjectionParamValue(exp, pos)
		if err != nil {
			return nil, err
		}
		params[name] = value
		ignoreWhitespaces(exp, pos)
		c := exp[*pos]
		if c == ')' {
			break
		} else if c == ',' {
			*pos++
		} else {
			return nil, fmt.Errorf("looking for `,' or ')' at char %d", *pos)
		}
	}
	return params, nil
}

// scanProjectionFieldName captures a field name at current position and advance
// the cursor position "pos" at the next character following the field name.
func scanProjectionFieldName(exp string, pos *int) string {
	ignoreWhitespaces(exp, pos)
	field := []byte{}
	for *pos < len(exp) {
		c := exp[*pos]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			field = append(field, c)
			*pos++
			continue
		}
		break
	}
	return string(field)
}

// scanProjectionFieldNameWithAlias parses a field optional alias followed by it's
// name separated by a column at current position and advance the cursor
// position "pos" at the next character following the field name.
func scanProjectionFieldNameWithAlias(exp string, pos *int) (name string, alias string) {
	name = scanProjectionFieldName(exp, pos)
	ignoreWhitespaces(exp, pos)
	if *pos < len(exp) && exp[*pos] == ':' {
		*pos++
		ignoreWhitespaces(exp, pos)
		alias = name
		name = scanProjectionFieldName(exp, pos)
	}
	return name, alias
}

// scanProjectionParamValue captures a parameter value at the current position and
// advance the cursor position "pos" at the next character following the field name.
//
// The returned value may be either a string if the value was quotted or a float
// if not an was a valid number. In case of syntax error, an error is returned.
func scanProjectionParamValue(exp string, pos *int) (interface{}, error) {
	ignoreWhitespaces(exp, pos)
	c := exp[*pos]
	if c == '"' || c == '\'' {
		quote := c
		quotted := false
		closed := false
		value := []byte{}
		*pos++
	L:
		for *pos < len(exp) {
			c := exp[*pos]
			if quotted {
				quotted = false
				value = append(value, c)
			} else {
				switch c {
				case '\\':
					quotted = true
				case quote:
					*pos++
					closed = true
					break L
				default:
					value = append(value, c)
				}
			}
			*pos++
		}
		if !closed {
			return nil, fmt.Errorf("looking for %c at char %d", quote, *pos)
		}
		return string(value), nil
	} else if (c >= '0' && c <= '9') || c == '-' {
		dot := false
		value := []byte{c}
		*pos++
		for *pos < len(exp) {
			c := exp[*pos]
			if c >= '0' && c <= '9' {
				value = append(value, c)
			} else if !dot && c == '.' {
				dot = true
				value = append(value, c)
			} else {
				break
			}
			*pos++
		}
		return strconv.ParseFloat(string(value), 64)
	} else {
		return nil, fmt.Errorf("looking for value at char %d", *pos)
	}
}

// ignoreWhitespaces advance the cursor position pos until non printable
// characters are met.
func ignoreWhitespaces(exp string, pos *int) {
	for *pos < len(exp) {
		c := exp[*pos]
		switch c {
		case ' ', '\n', '\r', '\t':
			// ignore witespaces
			*pos++
			continue
		}
		break
	}
}
