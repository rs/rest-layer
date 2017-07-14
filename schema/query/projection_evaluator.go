package query

import (
	"context"
	"fmt"

	"github.com/rs/rest-layer/schema"
)

// Eval evaluate the projection on the given payload with the help of the
// validator. The resolver is used to fetch payload of references outside of the
// provided payload.
func (p Projection) Eval(ctx context.Context, payload map[string]interface{}, validator schema.Validator, resolver ReferenceResolver) (map[string]interface{}, error) {
	payload, err := evalProjection(ctx, p, payload, validator, resolver)
	if err == nil {
		// The resulting payload may contain some asyncProjection, we must execute
		// them concurrently until there's no more.
		err = resolveAsyncProjections(ctx, payload, resolver)
	}
	return payload, err
}

func evalProjection(ctx context.Context, p Projection, payload map[string]interface{}, validator schema.Validator, resolver ReferenceResolver) (map[string]interface{}, error) {
	res := map[string]interface{}{}
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
					if subval, err = evalProjection(ctx, pf.Children, subval, def.Schema, resolver); err != nil {
						return nil, fmt.Errorf("%s.%v", pf.Name, err)
					}
					if res[name], err = resolveFieldHandler(ctx, pf, def, subval); err != nil {
						return nil, err
					}
				} else if ref, ok := def.Validator.(*schema.Reference); ok {
					// Do not execute the sub-request right away, store a
					// asyncProjection type of lambda that will be executed later
					// with concurrency control
					res[name] = asyncGet{
						payload: res,
						field:   pf,
						path:    ref.Path,
						id:      val,
						def:     def,
					}
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
				// Do not execute the sub-request right away, store a
				// asyncProjection type of lambda that will be executed later with
				// concurrency control
				handler, err := subResourceHandler(pf, def, ref.Path, resolver)
				if err != nil {
					return nil, err
				}
				res[name] = handler
			}
		}
	}
	return res, nil
}

type asyncProjection func(ctx context.Context) (interface{}, error)

type asyncGet struct {
	payload map[string]interface{}
	field   ProjectionField
	path    string
	id      interface{}
	def     *schema.Field
}

type asyncGetHandler func(ctx context.Context, payload map[string]interface{}, validator schema.Validator) (interface{}, error)

func subResourceHandler(pf ProjectionField, def *schema.Field, path string, resolver ReferenceResolver) (asyncProjection, error) {
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
	return func(ctx context.Context) (interface{}, error) {
		list, validator, err := resolver(ctx, path, q)
		if err != nil {
			return nil, fmt.Errorf("%s: error fetching sub-resource: %v", pf.Name, err)
		}
		vals := []interface{}{}
		for i, payload := range list {
			val, err := evalProjection(ctx, pf.Children, payload, validator, resolver)
			if err != nil {
				return nil, fmt.Errorf("%s: error applying projection on sub-resource item #%d: %v", pf.Name, i, err)
			}
			vals = append(vals, val)
		}
		return resolveFieldHandler(ctx, pf, def, vals)
	}, nil
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

func resolveAsyncProjections(ctx context.Context, p map[string]interface{}, resolver ReferenceResolver) error {
	for {
		sr := getProjectionResolvers(p)
		if len(sr) == 0 {
			break
		}
		done := make(chan error, len(sr))
		// TODO limit the number of sub requests.
		for _, r := range sr {
			go r(ctx, resolver, done)
		}
		wait := len(sr)
		cleanup := func() {
			// Make sure we empty the channel of remaining future responses
			// to prevent leaks.
			for wait > 0 {
				<-done
				wait--
			}
		}
		for wait > 0 {
			select {
			case err := <-done:
				wait--
				if err != nil {
					if wait > 0 {
						go cleanup()
					}
					return err
				}
			case <-ctx.Done():
				if wait > 0 {
					go cleanup()
				}
				return ctx.Err()
			}
		}
	}
	return nil
}

type asyncReferenceResolver func(ctx context.Context, resolver ReferenceResolver, done chan<- error)

func getProjectionResolvers(p map[string]interface{}) []asyncReferenceResolver {
	return append(getasyncReferenceResolvers(p), getAsyncGetResolver(p)...)
}

// getasyncReferenceResolvers parse the payload searching for any unresolved
// asyncProjection and build an asyncReferenceResolver for each ones.
func getasyncReferenceResolvers(p map[string]interface{}) []asyncReferenceResolver {
	as := []asyncReferenceResolver{}
	for name, val := range p {
		switch val := val.(type) {
		case asyncProjection:
			n := name
			as = append(as, func(ctx context.Context, resolver ReferenceResolver, done chan<- error) {
				res, err := val(ctx)
				if err == nil {
					p[n] = res
				}
				done <- err
			})
		case map[string]interface{}:
			as = append(as, getasyncReferenceResolvers(val)...)
		case []map[string]interface{}:
			for _, sval := range val {
				as = append(as, getasyncReferenceResolvers(sval)...)
			}
		}
	}
	return as
}

// getAsyncGetResolver search for any unresolved asyncGet and build on
// asyncReferenceResolver per resource with all requested ids coalesced.
func getAsyncGetResolver(p map[string]interface{}) []asyncReferenceResolver {
	ags := findAsyncGets(p)
	if len(ags) == 0 {
		return nil
	}
	// map of refPath -> []asyncGet.
	r := map[string][]asyncGet{}
	for _, ag := range ags {
		if _ags, found := r[ag.path]; found {
			r[ag.path] = append(_ags, ag)
		} else {
			r[ag.path] = []asyncGet{ag}
		}
	}
	as := make([]asyncReferenceResolver, 0, len(r))
	// create a resource resolver for each resource.
	for path, ags := range r {
		as = append(as, func(ctx context.Context, resolver ReferenceResolver, done chan<- error) {
			// Gather ids for each asyncGet
			q := &Query{}
			if len(ags) == 1 {
				q.Predicate = Predicate{
					Equal{Field: "id", Value: ags[0].id},
				}
			} else {
				ids := make([]Value, len(ags))
				for i, ag := range ags {
					ids[i] = ag.id
				}
				q.Predicate = Predicate{
					In{Field: "id", Values: ids},
				}
			}
			// Perform the resolution.
			payloads, validator, err := resolver(ctx, path, q)
			if err != nil {
				done <- err
				return
			}
			// Route back the value to corresponding asyncGet handlers.
			for _, ag := range ags {
				// Find the payload for this id.
				var payload map[string]interface{}
				for _, p := range payloads {
					// XXX we should not rely on the value of "id" to be equal to the requested id.
					if p["id"] == ag.id {
						payload = p
						break
					}
				}
				payload, err := evalProjection(ctx, ag.field.Children, payload, validator, resolver)
				if err != nil {
					done <- fmt.Errorf("%s: error applying Projection on sub-field: %v", ag.field.Name, err)
					return
				}
				val, err := resolveFieldHandler(ctx, ag.field, ag.def, payload)
				if err != nil {
					done <- err
					return
				}
				// Put the response value in place.
				ag.payload[ag.field.Name] = val
			}
			done <- nil
		})
	}
	return as
}

func findAsyncGets(p map[string]interface{}) []asyncGet {
	ag := []asyncGet{}
	for _, val := range p {
		switch val := val.(type) {
		case asyncGet:
			ag = append(ag, val)
		case map[string]interface{}:
			ag = append(ag, findAsyncGets(val)...)
		case []map[string]interface{}:
			for _, sval := range val {
				ag = append(ag, findAsyncGets(sval)...)
			}
		}
	}
	return ag
}
