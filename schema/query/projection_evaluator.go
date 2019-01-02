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
}

// Eval evaluate the projection on the given payload with the help of the
// validator. The resolver is used to fetch payload of references outside of the
// provided payload.
func (p Projection) Eval(ctx context.Context, payload map[string]interface{}, rsc Resource) (map[string]interface{}, error) {
	rbr := &referenceBatchResolver{}
	validator := rsc.Validator()
	payload, err := evalProjection(ctx, p, payload, validator, rbr)
	if err == nil {
		// Execute the batched reference resolutions.
		err = rbr.execute(ctx, rsc)
	}
	return payload, err
}

func evalProjection(ctx context.Context, p Projection, payload map[string]interface{}, validator schema.Validator, rbr *referenceBatchResolver) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	resMu := sync.Mutex{} // XXX use sync.Map once Go 1.9 is out
	if len(p) == 0 {
		// When the Projection is empty, it's like saying "all fields".
		// This allows notations like id,user{} to embed all fields of the user
		// sub-resource.
		for fn := range payload {
			p = append(p, ProjectionField{Name: fn})
		}
	}
	for i := range p {
		pf := p[i]
		name := pf.Name
		// Handle aliasing
		if pf.Alias != "" {
			name = pf.Alias
		}
		def := validator.GetField(pf.Name)
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
					if subval, err = evalProjection(ctx, pf.Children, subval, def.Schema, rbr); err != nil {
						return nil, fmt.Errorf("%s.%v", pf.Name, err)
					}
					if res[name], err = resolveFieldHandler(ctx, pf, def, subval); err != nil {
						return nil, err
					}
				} else if ref, ok := def.Validator.(*schema.Reference); ok {
					// Execute sub-request in batch
					var e Expression
					v, isArray := val.([]interface{})
					if isArray {
						e = &In{Field: "id", Values: v}
					} else {
						e = &Equal{Field: "id", Value: val}
					}
					q := &Query{
						Projection: pf.Children,
						Predicate:  Predicate{e},
					}
					rbr.request(ref.Path, q, func(payloads []map[string]interface{}, validator schema.Validator) error {
						var v interface{}
						if len(payloads) == 1 && isArray == false {
							if payloads[0] != nil {
								payload, err := evalProjection(ctx, pf.Children, payloads[0], validator, rbr)
								if err != nil {
									return fmt.Errorf("%s: error applying Projection on sub-field: %v", name, err)
								}
								if v, err = resolveFieldHandler(ctx, pf, def, payload); err != nil {
									return fmt.Errorf("%s: error resolving field handler on sub-field: %v", name, err)
								}
							}
						} else {
							var err error
							for i := range payloads {
								if payloads[i], err = evalProjection(ctx, pf.Children, payloads[i], validator, rbr); err != nil {
									return fmt.Errorf("%s: error applying projection on sub-field item #%d: %v", pf.Name, i, err)
								}
							}
							if v, err = resolveFieldHandler(ctx, pf, def, payloads); err != nil {
								return fmt.Errorf("%s: error resolving field handler on sub-field: %v", name, err)
							}
						}
						resMu.Lock()
						res[name] = v
						resMu.Unlock()
						return nil
					})
				} else {
					return nil, fmt.Errorf("%s: field as no children", pf.Name)
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
				rbr.request(ref.Path, q, func(payloads []map[string]interface{}, validator schema.Validator) (err error) {
					for i := range payloads {
						if payloads[i], err = evalProjection(ctx, pf.Children, payloads[i], validator, rbr); err != nil {
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
