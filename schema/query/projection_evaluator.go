package query

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/rest-layer/schema"
)

// Resource represents type that can be queried by Projection.Eval.
type Resource interface {
	// Find executes the query and returns the matching items.
	Find(ctx context.Context, query *Query) ([]map[string]interface{}, error)

	// MultiGet get some items by their id and return them in the same order. If one
	// or more item(s) is not found, their slot in the response is set to nil.
	MultiGet(ctx context.Context, ids []interface{}) ([]map[string]interface{}, error)

	// SubResource returns the sub-resource at path. If path starts with a
	// dot, the lookup is performed relative to the current resource.
	SubResource(ctx context.Context, path string) (Resource, error)

	// Validator returns the schema.Validator associated with the resource.
	Validator() schema.Validator

	// Path returns the full path of the resource composed of names of each
	// intermediate resources separated by dots (i.e.: res1.res2.res3).
	Path() string
}

// Eval evaluate the projection on the given payload with the help of the
// validator. The resolver is used to fetch payload of references outside of the
// provided payload.
func (p Projection) Eval(ctx context.Context, payload map[string]interface{}, rsc Resource) (map[string]interface{}, error) {
	rbr := &referenceBatchResolver{}
	validator := rsc.Validator()
	payload, err := evalProjection(ctx, p, payload, validator, rbr, rsc)
	if err == nil {
		// Execute the batched reference resolutions.
		err = rbr.execute(ctx)
	}
	return payload, err
}

func prepareProjection(p Projection, payload map[string]interface{}) (Projection, error) {
	var proj Projection
	if len(p) == 0 {
		// When the Projection is empty, it's like saying "all fields".
		// This allows notations like id,user{} to embed all fields of the user
		// sub-resource.
		for fn := range payload {
			proj = append(proj, ProjectionField{Name: fn})
		}
		return proj, nil
	}

	hasStar := false
	var starChildren Projection
	for _, pf := range p {
		if pf.Name == "*" {
			if hasStar {
				return nil, fmt.Errorf("only one * in projection allowed")
			}
			hasStar = true
			starChildren = pf.Children
		} else {
			proj = append(proj, pf)
		}
	}
	if hasStar {
		for fn := range payload {
			exists := false
			for _, pf := range proj {
				if fn == pf.Name && pf.Alias == "" {
					exists = true
				}
			}
			if !exists {
				proj = append(proj, ProjectionField{Name: fn, Children: starChildren})
			}
		}
	}
	return proj, nil
}

func evalProjectionArray(ctx context.Context, pf ProjectionField, payload []interface{}, def *schema.Field, rbr *referenceBatchResolver, rsc Resource) (*[]interface{}, error) {
	res := make([]interface{}, 0, len(payload))
	// Return pointer to res, because it may be populated after this function ends, by referenceBatchResolver
	// in `schema.Reference` case
	resp := &res
	resMu := sync.Mutex{}

	validator := def.Validator
	name := pf.Name
	if pf.Alias != "" {
		name = pf.Alias
	}

	switch fieldType := validator.(type) {
	// schema.Reference has higher priority than schema.FieldGetter
	case *schema.Reference:
		// Execute sub-request in batch
		e := &In{Field: "id", Values: payload}
		q := &Query{
			Projection: pf.Children,
			Predicate:  Predicate{e},
		}
		subRsc, err := rsc.SubResource(ctx, fieldType.Path)
		if err != nil {
			return nil, err
		}
		rbr.request(subRsc, q, func(payloads []map[string]interface{}, validator schema.Validator, rsc Resource) error {
			var v interface{}
			var err error
			for i := range payloads {
				if payloads[i], err = evalProjection(ctx, pf.Children, payloads[i], validator, rbr, rsc); err != nil {
					return fmt.Errorf("%s: error applying projection on sub-field item #%d: %v", pf.Name, i, err)
				}
			}
			if v, err = resolveFieldHandler(ctx, pf, def, payloads); err != nil {
				return fmt.Errorf("%s: error resolving field handler on sub-field: %v", name, err)
			}
			vv := v.([]map[string]interface{})
			resMu.Lock()
			for _, item := range vv {
				res = append(res, item)
			}
			resMu.Unlock()
			return nil
		})
	// schema.Array has higher priority than schema.FieldGetter
	case *schema.Array:
		for i, val := range payload {
			if subval, ok := val.([]interface{}); ok {
				var err error
				var subvalp *[]interface{}
				if subvalp, err = evalProjectionArray(ctx, pf, subval, &fieldType.Values, rbr, rsc); err != nil {
					return nil, fmt.Errorf("%s: error applying projection on array item #%d: %v", pf.Name, i, err)
				}
				var v interface{}
				if v, err = resolveFieldHandler(ctx, pf, def, *subvalp); err != nil {
					return nil, fmt.Errorf("%s: error resolving field handler on array: %v", name, err)
				}
				res = append(res, v)
			} else {
				return nil, fmt.Errorf("%s. is not an array", pf.Name)
			}
		}
	case schema.FieldGetter:
		for _, val := range payload {
			subval, ok := val.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("%s: invalid value: not a dict/object", pf.Name)
			}
			var err error
			if subval, err = evalProjection(ctx, pf.Children, subval, fieldType, rbr, rsc); err != nil {
				return nil, fmt.Errorf("%s.%v", pf.Name, err)
			}
			var v interface{}
			if v, err = resolveFieldHandler(ctx, pf, def, subval); err != nil {
				return nil, err
			}
			res = append(res, v)
		}
	default:
		return nil, fmt.Errorf("%s. unknown field type", pf.Name)
	}

	return resp, nil
}

func evalProjection(ctx context.Context, p Projection, payload map[string]interface{}, fg schema.FieldGetter, rbr *referenceBatchResolver, rsc Resource) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	resMu := sync.Mutex{}
	var err error
	p, err = prepareProjection(p, payload)
	if err != nil {
		return nil, err
	}
	for i := range p {
		pf := p[i]
		name := pf.Name
		// Handle aliasing
		if pf.Alias != "" {
			name = pf.Alias
		}
		def := fg.GetField(pf.Name)
		// Skip hidden fields
		if def != nil && def.Hidden {
			continue
		}
		if val, found := payload[pf.Name]; found {
			// Handle sub field selection (if field has a value)
			if len(pf.Children) > 0 && val != nil {
				if def != nil && def.Schema != nil {
					subval, ok := val.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("%s: invalid value: not a dict", pf.Name)
					}
					var err error
					if subval, err = evalProjection(ctx, pf.Children, subval, def.Schema, rbr, rsc); err != nil {
						return nil, fmt.Errorf("%s.%v", pf.Name, err)
					}
					if res[name], err = resolveFieldHandler(ctx, pf, def, subval); err != nil {
						return nil, err
					}
				} else if ref, ok := def.Validator.(*schema.Reference); ok {
					// Execute sub-request in batch
					q := &Query{
						Projection: pf.Children,
						Predicate:  Predicate{&Equal{Field: "id", Value: val}},
					}
					subRsc, err := rsc.SubResource(ctx, ref.Path)
					if err != nil {
						return nil, err
					}
					rbr.request(subRsc, q, func(payloads []map[string]interface{}, validator schema.Validator, rsc Resource) error {
						var v interface{}
						if len(payloads) == 1 {
							payload, err := evalProjection(ctx, pf.Children, payloads[0], validator, rbr, rsc)
							if err != nil {
								return fmt.Errorf("%s: error applying Projection on sub-field: %v", name, err)
							}
							if v, err = resolveFieldHandler(ctx, pf, def, payload); err != nil {
								return fmt.Errorf("%s: error resolving field handler on sub-field: %v", name, err)
							}
						}
						resMu.Lock()
						res[name] = v
						resMu.Unlock()
						return nil
					})
				} else if array, ok := def.Validator.(*schema.Array); ok {
					if payload, ok := val.([]interface{}); ok {
						var err error
						var subvalp *[]interface{}
						if subvalp, err = evalProjectionArray(ctx, pf, payload, &array.Values, rbr, rsc); err != nil {
							return nil, fmt.Errorf("%s: error applying projection on array item #%d: %v", pf.Name, i, err)
						}
						if res[name], err = resolveFieldHandler(ctx, pf, &array.Values, subvalp); err != nil {
							return nil, fmt.Errorf("%s: error resolving field handler on array: %v", name, err)
						}
					} else {
						return nil, fmt.Errorf("%s: invalid value: not an array", pf.Name)
					}
				} else if fg, ok := def.Validator.(schema.FieldGetter); ok {
					subval, ok := val.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("%s: invalid value: not a dict", pf.Name)
					}
					var err error
					if subval, err = evalProjection(ctx, pf.Children, subval, fg, rbr, rsc); err != nil {
						return nil, fmt.Errorf("%s.%v", pf.Name, err)
					}
					if res[name], err = resolveFieldHandler(ctx, pf, def, subval); err != nil {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf("%s: field has no children", pf.Name)
				}
			} else {
				var err error
				if res[name], err = resolveFieldHandler(ctx, pf, def, val); err != nil {
					return nil, err
				}
			}
		} else if def != nil {
			// If field is not found, it may be a connection
			if ref, ok := def.Validator.(*schema.Connection); ok {
				id, ok := payload["id"]
				if !ok {
					return nil, fmt.Errorf("%s: error applying projection on sub-resource: item lacks ID field", pf.Name)
				}
				q, err := connectionQuery(pf, ref.Field, id, ref.Validator)
				if err != nil {
					return nil, err
				}
				subRsc, err := rsc.SubResource(ctx, ref.Path)
				if err != nil {
					return nil, err
				}
				rbr.request(subRsc, q, func(payloads []map[string]interface{}, validator schema.Validator, rsc Resource) (err error) {
					for i := range payloads {
						if payloads[i], err = evalProjection(ctx, pf.Children, payloads[i], validator, rbr, rsc); err != nil {
							return fmt.Errorf("%s: error applying projection on sub-resource item #%d: %v", pf.Name, i, err)
						}
					}
					var v interface{}
					if v, err = resolveFieldHandler(ctx, pf, def, payloads); err != nil {
						return fmt.Errorf("%s: error resolving field handler on sub-resource: %v", name, err)
					}
					resMu.Lock()
					res[name] = v
					resMu.Unlock()
					return nil
				})
			}
		}
	}
	return res, nil
}

// connectionQuery builds a query from a projection field on a schema.Connection type field.
func connectionQuery(pf ProjectionField, field string, id interface{}, validator schema.Validator) (*Query, error) {
	q := &Query{
		Projection: pf.Children,
		Predicate:  Predicate{&Equal{Field: field, Value: id}},
	}
	if filter, ok := pf.Params["filter"].(string); ok {
		p, err := ParsePredicate(filter)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid filter: %v", pf.Name, err)
		}
		err = p.Prepare(validator)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid filter: %v", pf.Name, err)
		}
		q.Predicate = append(q.Predicate, p...)
	}
	if sort, ok := pf.Params["sort"].(string); ok {
		s, err := ParseSort(sort)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid sort: %v", pf.Name, err)
		}
		q.Sort = s
	}
	skip := 0
	if v, ok := pf.Params["skip"].(int); ok {
		skip = v
	}
	page := 1
	if v, ok := pf.Params["page"].(int); ok {
		page = v
	}
	limit := 20
	if v, ok := pf.Params["limit"].(int); ok {
		limit = v
	}
	q.Window = Page(page, limit, skip)
	return q, nil
}

// resolveFieldHandler calls the field handler with the provided params (if any).
func resolveFieldHandler(ctx context.Context, pf ProjectionField, def *schema.Field, val interface{}) (interface{}, error) {
	if def == nil {
		return val, nil
	}
	var err error
	if def.Handler != nil && len(pf.Params) > 0 {
		val, err = def.Handler(ctx, val, pf.Params)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", pf.Name, err)
		}
	}
	if s, ok := def.Validator.(schema.FieldSerializer); ok {
		val, err = s.Serialize(val)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", pf.Name, err)
		}
	}
	return val, nil
}
