package resource

import (
	"fmt"
	"strconv"

	"github.com/rs/rest-layer/schema"
)

/*
Selector expression syntax allows to list fields that must be kept in the response
hierarchically.

A field is an alphanum + - and _ separated by comas:

field1,field2

When a document has sub-fields, sub-resources or sub-connections, the sub-element's
fields can be specified as well by enclosing them between brackets:

field1{sub-field1,sub-field2},field2

Fields can get some some parameters which can be passed to field filters to transform
the value. Parameters are passed as key=value pairs enclosed in parenthezies, with value
being either a quotted string or a numerical value:

field1(param1="value", param2=123),field2

You can combine field params and sub-field definition:

field1(param1="value", param2=123){sub-field1,sub-field2},field2

Or pass params to sub-fields:

field1{sub-field1(param1="value"),sub-field2},field2

Fields can also be renamed (aliased). This is useful when you want to have several times
the same fields with different sets of parameters. To define alias, follow the field
definition by a colon (:) followed by the alias:

field:alias

With params:

thumbnail_url(size=80):thumbnail_small_url,thumbnail_url(size=500):thumbnail_large_url

With this example, the resulted document would be:

  {
    "thumbnail_small_url": "the url with size 80",
    "thumbnail_large_url": "the url with size 500",
  }

*/

func parseSelectorExpression(s []byte, pos *int, ln int, opened bool) ([]Field, error) {
	selector := []Field{}
	var field *Field
	for *pos < ln {
		if field == nil {
			name := scanSelectorFieldName(s, pos, ln)
			if name == "" {
				return nil, fmt.Errorf("looking for field name at char %d", *pos)
			}
			field = &Field{Name: name}
			continue
		}
		c := s[*pos]
		switch c {
		case '{':
			if field.Alias != "" {
				return nil, fmt.Errorf("looking for `,` and got `{' at char %d", *pos)
			}
			*pos++
			flds, err := parseSelectorExpression(s, pos, ln, true)
			if err != nil {
				return nil, err
			}
			field.Fields = flds
		case '}':
			if opened {
				selector = append(selector, *field)
				return selector, nil
			}
			return nil, fmt.Errorf("looking for field name and got `}' at char %d", *pos)
		case '(':
			*pos++
			params, err := parseSelectorFieldParams(s, pos, ln)
			if err != nil {
				return nil, err
			}
			field.Params = params
		case ':':
			*pos++
			name := scanSelectorFieldName(s, pos, ln)
			if name == "" {
				return nil, fmt.Errorf("looking for field alias at char %d", *pos)
			}
			field.Alias = name
			continue
		case ',':
			selector = append(selector, *field)
			field = nil
		case ' ', '\n', '\r', '\t':
			// ignore witespaces
		default:
			return nil, fmt.Errorf("invalid char at %d", *pos)
		}
		*pos++
	}
	if opened {
		return nil, fmt.Errorf("looking for `}' at char %d", *pos)
	}
	if field != nil {
		selector = append(selector, *field)
	}
	return selector, nil
}

func parseSelectorFieldParams(s []byte, pos *int, ln int) (map[string]interface{}, error) {
	params := map[string]interface{}{}
	for *pos < ln {
		name := scanSelectorFieldName(s, pos, ln)
		if name == "" {
			return nil, fmt.Errorf("looking for parameter name at char %d", *pos)
		}
		found := false
	L:
		for *pos < ln {
			c := s[*pos]
			switch c {
			case '=':
				found = true
				break L
			case ' ', '\n', '\r', '\t':
				// ignore whitespaces
			default:
				return nil, fmt.Errorf("looking for = at char %d", *pos)
			}
			*pos++
		}
		if !found {
			return nil, fmt.Errorf("looking for = at char %d", *pos)
		}
		*pos++
		value, err := scanSelectorValue(s, pos, ln)
		if err != nil {
			return nil, err
		}
		params[name] = value
		ignoreWhitespaces(s, pos, ln)
		c := s[*pos]
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

func scanSelectorFieldName(s []byte, pos *int, ln int) string {
	ignoreWhitespaces(s, pos, ln)
	field := []byte{}
	for *pos < ln {
		c := s[*pos]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			field = append(field, c)
			*pos++
			continue
		}
		break
	}
	return string(field)
}

func scanSelectorValue(s []byte, pos *int, ln int) (interface{}, error) {
	ignoreWhitespaces(s, pos, ln)
	c := s[*pos]
	if c == '"' || c == '\'' {
		quote := c
		quotted := false
		closed := false
		value := []byte{}
		*pos++
	L:
		for *pos < ln {
			c := s[*pos]
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
		for *pos < ln {
			c := s[*pos]
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

func ignoreWhitespaces(s []byte, pos *int, ln int) {
	for *pos < ln {
		c := s[*pos]
		switch c {
		case ' ', '\n', '\r', '\t':
			// ignore witespaces
			*pos++
			continue
		}
		break
	}
}

func validateSelector(s []Field, r *Resource, v schema.Validator) error {
	for _, f := range s {
		def := v.GetField(f.Name)
		if def == nil {
			return fmt.Errorf("%s: unknown field", f.Name)
		}
		if len(f.Fields) > 0 {
			if def.Schema == nil {
				return fmt.Errorf("%s: field as no children", f.Name)
			}
			if err := validateSelector(f.Fields, r, def.Schema); err != nil {
				return fmt.Errorf("%s.%s", f.Name, err.Error())
			}
		}
		// TODO: support references
		// TODO: support connections
		// TODO: support params
		if len(f.Params) > 0 {
			return fmt.Errorf("%s: params are not yet supported", f.Name)
		}
	}
	return nil
}

func applySelector(s []Field, p map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for _, f := range s {
		if val, found := p[f.Name]; found {
			name := f.Name
			if f.Alias != "" {
				name = f.Alias
			}
			if len(f.Fields) > 0 {
				res[name] = applySelector(f.Fields, val.(map[string]interface{}))
			} else {
				res[name] = val
			}
		}
	}
	return res
}
