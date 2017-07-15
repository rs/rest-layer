package query

import (
	"context"
	"sync"

	"github.com/rs/rest-layer/schema"
)

type referenceResponseHandler func(payloads []map[string]interface{}, validator schema.Validator) error

type referenceBatchResolver struct {
	mu       sync.Mutex
	requests []referenceRequest
}

func (rbr *referenceBatchResolver) request(resourcePath string, q *Query, handler referenceResponseHandler) {
	rbr.mu.Lock()
	defer rbr.mu.Unlock()
	if len(q.Predicate) == 1 {
		if eq, ok := q.Predicate[0].(Equal); ok && eq.Field == "id" {
			// Make an optimization for query on a single id so we can coalese them into a single request.
			id := eq.Value
			for _, r := range rbr.requests {
				if r, ok := r.(*referenceMultiGetRequest); ok && r.resourcePath == resourcePath {
					r.add(id, handler)
					return
				}
			}
			// Not found an existing multi get request for this path, create a new one.
			r := &referenceMultiGetRequest{resourcePath: resourcePath}
			r.add(id, handler)
			rbr.requests = append(rbr.requests, r)
			return
		}
	}
	rbr.requests = append(rbr.requests, referenceSingleRequest{
		resourcePath: resourcePath,
		query:        q,
		handler:      handler,
	})
}

func (rbr *referenceBatchResolver) execute(ctx context.Context, resolver ReferenceResolver) error {
	for len(rbr.requests) > 0 {
		// Get the list of requests.
		requests := rbr.requests
		// Reset the request queue so sub-request can append new ones.
		rbr.requests = []referenceRequest{}
		// Execute the requests in parallel.
		wg := &sync.WaitGroup{}
		wg.Add(len(requests))
		var err error
		for i := range requests {
			r := requests[i]
			go func() {
				if e := r.execute(ctx, resolver); e != nil {
					err = e
				}
				wg.Done()
			}()
		}
		wg.Wait()
		if err != nil {
			return err
		}
		// If sub-requests scheduled new request, loop and execute them.
	}
	return nil
}

type referenceRequest interface {
	execute(ctx context.Context, resolver ReferenceResolver) error
}

type referenceSingleRequest struct {
	resourcePath string
	query        *Query
	handler      referenceResponseHandler
}

func (r referenceSingleRequest) execute(ctx context.Context, resolver ReferenceResolver) error {
	payloads, validator, err := resolver(ctx, r.resourcePath, r.query)
	if err != nil {
		return err
	}
	return r.handler(payloads, validator)
}

type referenceMultiGetRequest struct {
	resourcePath string
	ids          []Value
	handlers     []referenceResponseHandler
	done         []bool
}

func (r *referenceMultiGetRequest) add(id interface{}, handler referenceResponseHandler) {
	r.ids = append(r.ids, id)
	r.handlers = append(r.handlers, handler)
	r.done = append(r.done, false)
}

func (r *referenceMultiGetRequest) execute(ctx context.Context, resolver ReferenceResolver) error {
	q := &Query{}
	if len(r.ids) == 1 {
		q.Predicate = Predicate{
			Equal{Field: "id", Value: r.ids[0]},
		}
	} else {
		q.Predicate = Predicate{
			In{Field: "id", Values: r.ids},
		}
	}
	payloads, validator, err := resolver(ctx, r.resourcePath, q)
	if err != nil {
		return err
	}
	payloadsWrapper := make([]map[string]interface{}, 1)
	for _, p := range payloads {
		for i, id := range r.ids {
			// XXX we should not rely on the value of "id" to be equal to the requested id.
			if id == p["id"] {
				payloadsWrapper[0] = p
				handler := r.handlers[i]
				if err := handler(payloadsWrapper, validator); err != nil {
					return err
				}
				r.done[i] = true
			}
		}
	}
	// Call handlers for ids which did not get a matching response (i.e. not found).
	payloadsEmptyWrapper := payloadsWrapper[:0]
	for i, done := range r.done {
		if !done {
			handler := r.handlers[i]
			if err := handler(payloadsEmptyWrapper, validator); err != nil {
				return err
			}
		}
	}
	return nil
}
