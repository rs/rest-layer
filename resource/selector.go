package resource

import (
	"context"
	"fmt"

	"github.com/rs/rest-layer/schema"
)

type asyncSelector func(ctx context.Context) (interface{}, error)

type asyncGet struct {
	payload  map[string]interface{}
	field    string
	resource *Resource
	id       interface{}
	handler  func(ctx context.Context, item *Item) (interface{}, error)
}

func validateSelector(s []Field, v schema.Validator) error {
	for _, f := range s {
		def := v.GetField(f.Name)
		if def == nil {
			return fmt.Errorf("%s: unknown field", f.Name)
		}
		if def.Hidden {
			// Hidden fields can't be selected
			return fmt.Errorf("%s: hidden field", f.Name)
		}
		if len(f.Fields) > 0 {
			if def.Schema != nil {
				// Sub-field on a dict (sub-schema)
				if err := validateSelector(f.Fields, def.Schema); err != nil {
					return fmt.Errorf("%s.%v", f.Name, err)
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
			if len(def.Params) == 0 {
				return fmt.Errorf("%s: params not allowed", f.Name)
			}
			for name, value := range f.Params {
				param, found := def.Params[name]
				if !found {
					return fmt.Errorf("%s: unsupported param name: %s", f.Name, name)
				}
				if param.Validator != nil {
					var err error
					value, err = param.Validator.Validate(value)
					if err != nil {
						return fmt.Errorf("%s: invalid param `%s' value: %v", f.Name, name, err)
					}
				}
				f.Params[name] = value
			}
		}
	}
	return nil
}

func applySelector(ctx context.Context, s []Field, v schema.Validator, p map[string]interface{}, resolver ReferenceResolver) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	if len(s) == 0 {
		// When the field selector is empty, it's like saying "all fields".
		// This allows notations like id,user{} to embed all fields of the user
		// sub-resource.
		for fn := range p {
			s = append(s, Field{Name: fn})
		}
	}
	for _, f := range s {
		name := f.Name
		// Handle aliasing
		if f.Alias != "" {
			name = f.Alias
		}
		def := v.GetField(f.Name)
		// Skip hidden fields
		if def != nil && def.Hidden {
			continue
		}
		if val, found := p[f.Name]; found {
			// Handle sub field selection (if field has a value)
			if len(f.Fields) > 0 && val != nil {
				if def != nil && def.Schema != nil {
					subval, ok := val.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("%s: invalid value: not a dict", f.Name)
					}
					var err error
					if subval, err = applySelector(ctx, f.Fields, def.Schema, subval, resolver); err != nil {
						return nil, fmt.Errorf("%s.%v", f.Name, err)
					}
					if res[name], err = resolveFieldHandler(ctx, f, def, subval); err != nil {
						return nil, err
					}
				} else if ref, ok := def.Validator.(*schema.Reference); ok {
					// Sub-field on a reference (sub-request)
					rsrc, err := resolver(ref.Path)
					if err != nil {
						return nil, fmt.Errorf("%s: error linking sub-field resource: %v", f.Name, err)
					}
					// Do not execute the sub-request right away, store a asyncSelector type of
					// lambda that will be executed later with concurrency control
					res[name] = asyncGet{
						payload:  res,
						field:    name,
						resource: rsrc,
						id:       val,
						handler:  subFieldHandler(f, def, rsrc, resolver),
					}
				} else {
					return nil, fmt.Errorf("%s: field as no children", f.Name)
				}
			} else {
				var err error
				if res[name], err = resolveFieldHandler(ctx, f, def, val); err != nil {
					return nil, err
				}
			}
		} else if def != nil {
			// If field is not found, it may be a connection
			if ref, ok := def.Validator.(connection); ok {
				// Sub-field on a sub resource (sub-request)
				rsrc, err := resolver(ref.path)
				if err != nil {
					return nil, fmt.Errorf("%s: error linking sub-resource: %v", f.Name, err)
				}
				// Do not execute the sub-request right away, store a asyncSelector type of
				// lambda that will be executed later with concurrency control
				res[name] = asyncSelector(subResourceHandler(f, def, rsrc, resolver))
			}
		}
	}
	return res, nil
}

// resolveFieldHandler handles selector handler / params
func resolveFieldHandler(ctx context.Context, f Field, def *schema.Field, val interface{}) (interface{}, error) {
	if def == nil {
		return val, nil
	}
	var err error
	if def.Handler != nil && len(f.Params) > 0 {
		val, err = def.Handler(ctx, val, f.Params)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", f.Name, err)
		}
	}
	if s, ok := def.Validator.(schema.FieldSerializer); ok {
		val, err = s.Serialize(val)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", f.Name, err)
		}
	}
	return val, nil
}

func subFieldHandler(f Field, def *schema.Field, rsrc *Resource, resolver ReferenceResolver) func(ctx context.Context, item *Item) (interface{}, error) {
	return func(ctx context.Context, item *Item) (interface{}, error) {
		val, err := applySelector(ctx, f.Fields, rsrc.Validator(), item.Payload, resolver)
		if err != nil {
			return nil, fmt.Errorf("%s: error applying selector on sub-field: %v", f.Name, err)
		}
		return resolveFieldHandler(ctx, f, def, val)
	}
}

func subResourceHandler(f Field, def *schema.Field, rsrc *Resource, resolver ReferenceResolver) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		l := NewLookup()
		if filter, ok := f.Params["filter"].(string); ok {
			err := l.AddFilter(filter, rsrc.Validator())
			if err != nil {
				return nil, fmt.Errorf("%s: invalid filter: %v", f.Name, err)
			}
		}
		if sort, ok := f.Params["sort"].(string); ok {
			err := l.SetSort(sort, rsrc.Validator())
			if err != nil {
				return nil, fmt.Errorf("%s: invalid sort: %v", f.Name, err)
			}
		}
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
			return nil, fmt.Errorf("%s: error fetching sub-resource: %v", f.Name, err)
		}
		vals := []interface{}{}
		for i, item := range list.Items {
			var val interface{}
			var err error
			val, err = applySelector(ctx, f.Fields, rsrc.Validator(), item.Payload, resolver)
			if err != nil {
				return nil, fmt.Errorf("%s: error applying selector on sub-resource item #%d: %v", f.Name, i, err)
			}
			vals = append(vals, val)
		}
		return resolveFieldHandler(ctx, f, def, vals)
	}
}
