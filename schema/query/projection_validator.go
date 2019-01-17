package query

import (
	"fmt"

	"github.com/rs/rest-layer/schema"
)

// Validate validates the projection field against the provided validator.
func (pf ProjectionField) Validate(fg schema.FieldGetter) error {
	if pf.Name == "*" {
		if pf.Alias != "" {
			return fmt.Errorf("%s: can't have an alias", pf.Name)
		}
		return nil
	}

	def := fg.GetField(pf.Name)
	if def == nil {
		return fmt.Errorf("%s: unknown field", pf.Name)
	}
	if def.Hidden {
		// Hidden fields can't be selected
		return fmt.Errorf("%s: hidden field", pf.Name)
	}
	if len(pf.Children) > 0 {
		if def.Schema != nil {
			// Sub-field on a dict (sub-schema)
			if err := pf.Children.Validate(def.Schema); err != nil {
				return fmt.Errorf("%s.%v", pf.Name, err)
			}
		} else if ref, ok := def.Validator.(*schema.Reference); ok {
			// Sub-field on a reference (sub-request)
			if err := pf.Children.Validate(ref.SchemaValidator); err != nil {
				return fmt.Errorf("%s.%v", pf.Name, err)
			}
		} else if conn, ok := def.Validator.(*schema.Connection); ok {
			// Sub-field on a sub resource (sub-request)
			if err := pf.Children.Validate(conn.Validator); err != nil {
				return fmt.Errorf("%s.%v", pf.Name, err)
			}
		} else if _, ok := def.Validator.(*schema.Dict); ok {
			// Sub-field on a dict resource
		} else if array, ok := def.Validator.(*schema.Array); ok {
			if fg, ok := array.Values.Validator.(schema.FieldGetter); ok {
				if err := pf.Children.Validate(fg); err != nil {
					return fmt.Errorf("%s.%v", pf.Name, err)
				}
			}
		} else {
			return fmt.Errorf("%s: field has no children", pf.Name)
		}
	}
	if len(pf.Params) > 0 {
		if len(def.Params) == 0 {
			return fmt.Errorf("%s: params not allowed", pf.Name)
		}
		for name, value := range pf.Params {
			param, found := def.Params[name]
			if !found {
				return fmt.Errorf("%s: unsupported param name: %s", pf.Name, name)
			}
			if param.Validator != nil {
				var err error
				value, err = param.Validator.Validate(value)
				if err != nil {
					return fmt.Errorf("%s: invalid param `%s' value: %v", pf.Name, name, err)
				}
			}
			pf.Params[name] = value
		}
	}
	return nil
}
