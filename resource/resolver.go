package resource

import "golang.org/x/net/context"

func resolveAsyncSelectors(ctx context.Context, p map[string]interface{}) error {
	for {
		sr := getSelectorResolvers(p)
		if len(sr) == 0 {
			break
		}
		done := make(chan error, len(sr))
		// TODO limit the number of // sub requests
		for _, r := range sr {
			go r(ctx, done)
		}
		wait := len(sr)
		cleanup := func() {
			// Make sure we empty the channel of remaining future responses
			// to prevent leaks
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

type asyncSelectorResolver func(ctx context.Context, done chan<- error)

func getSelectorResolvers(p map[string]interface{}) []asyncSelectorResolver {
	return append(getAsyncSelectorResolvers(p), getAsyncGetResolver(p)...)
}

// getAsyncSelectorResolvers parse the payload searching for any unresolved asyncSelector
// and build an asyncSelectorResolver for each ones.
func getAsyncSelectorResolvers(p map[string]interface{}) []asyncSelectorResolver {
	as := []asyncSelectorResolver{}
	for name, val := range p {
		switch val := val.(type) {
		case asyncSelector:
			n := name
			as = append(as, func(ctx context.Context, done chan<- error) {
				res, err := val(ctx)
				if err == nil {
					p[n] = res
				}
				done <- err
			})
		case map[string]interface{}:
			as = append(as, getAsyncSelectorResolvers(val)...)
		case []map[string]interface{}:
			for _, sval := range val {
				as = append(as, getAsyncSelectorResolvers(sval)...)
			}
		}
	}
	return as
}

// getAsyncGetResolver search for any unresolved asyncGet and build on asyncSelectorResolver
// per resource with all requested ids coalesced.
func getAsyncGetResolver(p map[string]interface{}) []asyncSelectorResolver {
	ags := findAsyncGets(p)
	if len(ags) == 0 {
		return nil
	}
	// map of resource -> []asyncGet
	r := map[*Resource][]asyncGet{}
	for _, ag := range ags {
		if _ags, found := r[ag.resource]; found {
			r[ag.resource] = append(_ags, ag)
		} else {
			r[ag.resource] = []asyncGet{ag}
		}
	}
	as := make([]asyncSelectorResolver, 0, len(r))
	// create a resource resolver for each resource
	for rsrc, ags := range r {
		as = append(as, func(ctx context.Context, done chan<- error) {
			// Gater ids for each asyncGet
			ids := make([]interface{}, len(ags))
			for i, ag := range ags {
				ids[i] = ag.id
			}
			// Perform the mget
			items, err := rsrc.MultiGet(ctx, ids)
			if err != nil {
				done <- err
				return
			}
			// Route back the value to corresponding asyncGet handlers
			for i, ag := range ags {
				val, err := ag.handler(ctx, items[i])
				if err != nil {
					done <- err
					return
				}
				// Put the response value in place
				ag.payload[ag.field] = val
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
