package query

import (
	"context"
	"fmt"

	"sync"

	"github.com/rs/rest-layer/schema"
)

// Eval evaluate the projection on the given payload with the help of the
// validator. The resolver is used to fetch payload of references outside of the
// provided payload.
func (p Projection) Eval(ctx context.Context, payload map[string]interface{}, validator schema.Validator, resolver ReferenceResolver) (map[string]interface{}, error) {
	rbr := &referenceBatchResolver{}
	payload, err := evalProjection(ctx, p, payload, validator, rbr)
	if err == nil {
		// Execute the batched reference resolutions.
		err = rbr.execute(ctx, resolver)
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
	for _, pf := range p {
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
					q := &Query{Predicate: Predicate{Equal{Field: "id", Value: val}}}
					rbr.request(ref.Path, q, func(payloads []map[string]interface{}, validator schema.Validator) error {
						resMu.Lock()
						defer resMu.Unlock()
						if len(payloads) != 1 {
							res[name] = nil
							return nil
						}
						payload, err := evalProjection(ctx, pf.Children, payloads[0], validator, rbr)
						if err != nil {
							return fmt.Errorf("%s: error applying Projection on sub-field: %v", name, err)
						}
						if res[name], err = resolveFieldHandler(ctx, pf, def, payload); err != nil {
							return fmt.Errorf("%s: error resolving field handler on sub-field: %v", name, err)
						}
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
				q, err := connectionQuery(pf)
				if err != nil {
					return nil, err
				}
				rbr.request(ref.Path, q, func(payloads []map[string]interface{}, validator schema.Validator) (err error) {
					for i := range payloads {
						if payloads[i], err = evalProjection(ctx, pf.Children, payloads[i], validator, rbr); err != nil {
							return fmt.Errorf("%s: error applying projection on sub-resource item #%d: %v", pf.Name, i, err)
						}
					}
					resMu.Lock()
					defer resMu.Unlock()
					if res[name], err = resolveFieldHandler(ctx, pf, def, payloads); err != nil {
						return fmt.Errorf("%s: error resolving field handler on sub-resource: %v", name, err)
					}
					return nil
				})
			}
		}
	}
	return res, nil
}

// connectionQuery builds a query from a projection field on a schema.Connection type field.
func connectionQuery(pf ProjectionField) (*Query, error) {
	q := &Query{}
	if filter, ok := pf.Params["filter"].(string); ok {
		p, err := ParsePredicate(filter)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid filter: %v", pf.Name, err)
		}
		q.Predicate = p
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
