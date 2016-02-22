package resource

import (
	"fmt"

	"github.com/rs/rest-layer/schema"
)

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
