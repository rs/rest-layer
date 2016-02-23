package resource

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/rs/rest-layer/schema"
)

type asyncSelector func(ctx context.Context) (interface{}, error)

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
			} else if _, ok := def.Validator.(connection); ok {
				// Sub-field on a sub resource (sub-request)
			} else {
				return fmt.Errorf("%s: field as no children", f.Name)
			}
		}
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
		name := f.Name
		if val, found := p[name]; found {
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
					rsrc, err := resolver(ref.Path)
					if err != nil {
						return nil, fmt.Errorf("%s: error linking sub-field resource: %s", f.Name, err.Error())
					}
					// Do not execute the sub-request right away, store a asyncSelector type of
					// lambda that will be executed later with concurrency control
					res[name] = asyncSelector(func(ctx context.Context) (interface{}, error) {
						l := NewLookup()
						l.AddQuery(schema.Query{schema.Equal{Field: "id", Value: val}})
						list, err := rsrc.Find(ctx, l, 1, 1)
						if err != nil {
							return nil, fmt.Errorf("%s: error fetching sub-field resource: %s", f.Name, err.Error())
						}
						subval := map[string]interface{}{}
						if len(list.Items) > 0 {
							subval = list.Items[0].Payload
						}
						subval, err = applySelector(f.Fields, rsrc.Validator(), subval, resolver)
						if err != nil {
							return nil, fmt.Errorf("%s: error applying selector on sub-field: %s", f.Name, err.Error())
						}
						return subval, nil
					})
				} else {
					return nil, fmt.Errorf("%s: field as no children", f.Name)
				}
			} else {
				res[name] = val
			}
		} else if def := v.GetField(f.Name); def != nil {
			// If field is not found, it may be a connection
			if ref, ok := def.Validator.(connection); ok {
				// Sub-field on a sub resource (sub-request)
				rsrc, err := resolver(ref.path)
				if err != nil {
					return nil, fmt.Errorf("%s: error linking sub-resource: %s", f.Name, err.Error())
				}
				// Do not execute the sub-request right away, store a asyncSelector type of
				// lambda that will be executed later with concurrency control
				res[name] = asyncSelector(func(ctx context.Context) (interface{}, error) {
					l := NewLookup()
					// TODO: parse params to add query and sort
					page := 1
					if v, ok := f.Params["page"].(int); ok {
						page = v
					}
					perPage := 20
					if v, ok := f.Params["limit"].(int); ok {
						perPage = v
					}
					list, err := rsrc.Find(ctx, l, page, perPage)
					if err != nil {
						return nil, fmt.Errorf("%s: error fetching sub-resource: %s", f.Name, err.Error())
					}
					subvals := []map[string]interface{}{}
					for i, item := range list.Items {
						subval, err := applySelector(f.Fields, rsrc.Validator(), item.Payload, resolver)
						if err != nil {
							return nil, fmt.Errorf("%s: error applying selector on sub-resource item #%d: %s", f.Name, i, err.Error())
						}
						subvals = append(subvals, subval)
					}
					return subvals, nil
				})
			}
		}
	}
	return res, nil
}
