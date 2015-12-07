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

func validateSelector(s []Field, v schema.Validator) error {
	for _, f := range s {
		def := v.GetField(f.Name)
		if def == nil {
			return fmt.Errorf("%s: unknown field", f.Name)
		}
		if len(f.Fields) > 0 {
			if def.Schema != nil {
				// Sub-field on a dict (sub-schema)
				if err := validateSelector(f.Fields, def.Schema); err != nil {
					return fmt.Errorf("%s.%s", f.Name, err.Error())
				}
			} else if _, ok := def.Validator.(*schema.Reference); ok {
				// Sub-field on a reference (sub-request)
			} else {
				return fmt.Errorf("%s: field as no children", f.Name)
			}
		}
		// TODO: support connections
		if len(f.Params) > 0 {
			if def.Params == nil {
				return fmt.Errorf("%s: params not allowed", f.Name)
			}
			for param, value := range f.Params {
				val, found := def.Params.Validators[param]
				if !found {
					return fmt.Errorf("%s: unsupported param name: %s", f.Name, param)
				}
				value, err := val.Validate(value)
				if err != nil {
					return fmt.Errorf("%s: invalid param `%s' value: %s", f.Name, param, err.Error())
				}
				f.Params[param] = value
			}
		}
	}
	return nil
}

func applySelector(s []Field, v schema.Validator, p map[string]interface{}, resolver ReferenceResolver) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	for _, f := range s {
		if val, found := p[f.Name]; found {
			name := f.Name
			// Handle aliasing
			if f.Alias != "" {
				name = f.Alias
			}
			// Handle selector params
			if len(f.Params) > 0 {
				def := v.GetField(f.Name)
				if def == nil || def.Params == nil {
					return nil, fmt.Errorf("%s: params not allowed", f.Name)
				}
				var err error
				val, err = def.Params.Handler(val, f.Params)
				if err != nil {
					return nil, fmt.Errorf("%s: %s", f.Name, err.Error())
				}
			}
			// Handle sub field selection (if field has a value)
			if len(f.Fields) > 0 && val != nil {
				def := v.GetField(f.Name)
				if def != nil && def.Schema != nil {
					subval, ok := val.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("%s: invalid value: not a dict", f.Name)
					}
					var err error
					res[name], err = applySelector(f.Fields, def.Schema, subval, resolver)
					if err != nil {
						return nil, fmt.Errorf("%s.%s", f.Name, err.Error())
					}
				} else if ref, ok := def.Validator.(*schema.Reference); ok {
					// Sub-field on a reference (sub-request)
					subres, subval, err := resolver(ref.Path, val)
					if err != nil {
						return nil, fmt.Errorf("%s: error fetching sub-field: %s", f.Name, err.Error())
					}
					res[name], err = applySelector(f.Fields, subres.validator, subval, resolver)
				} else {
					return nil, fmt.Errorf("%s: field as no children", f.Name)
				}
			} else {
				res[name] = val
			}
		}
	}
	return res, nil
}
