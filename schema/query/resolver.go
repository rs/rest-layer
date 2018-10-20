package query

import (
	"context"
	"errors"
	"sync"

	"github.com/rs/rest-layer/schema"
)

type referenceResponseHandler func(payloads []map[string]interface{}, validator schema.Validator) error

type referenceBatchResolver struct {
	mu       sync.Mutex
	requests []referenceRequest
}

func (rbr *referenceBatchResolver) request(resourcePath string, q *Query, handler referenceResponseHandler) {
	if len(q.Predicate) == 1 {
		if eq, ok := q.Predicate[0].(*Equal); ok && eq.Field == "id" {
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
			rbr.appendRequest(r)
			return
		}
	}
	rbr.appendRequest(referenceSingleRequest{
		resourcePath: resourcePath,
		query:        q,
		handler:      handler,
	})
}

func (rbr *referenceBatchResolver) appendRequest(r referenceRequest) {
	rbr.mu.Lock()
	defer rbr.mu.Unlock()
	rbr.requests = append(rbr.requests, r)
}

func (rbr *referenceBatchResolver) execute(ctx context.Context, rsc Resource) error {
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
				if e := r.execute(ctx, rsc); e != nil {
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
	execute(ctx context.Context, rsc Resource) error
}

type referenceSingleRequest struct {
	resourcePath string
	query        *Query
	handler      referenceResponseHandler
}

func (r referenceSingleRequest) execute(ctx context.Context, rsc Resource) error {
	subRsc, err := rsc.SubResource(ctx, r.resourcePath)
	if err != nil {
		return err
	}
	payloads, err := subRsc.Find(ctx, r.query)
	if err != nil {
		return err
	}
	return r.handler(payloads, subRsc.Validator())
}

type referenceMultiGetRequest struct {
	resourcePath string
	ids          []interface{}
	handlers     []referenceResponseHandler
}

func (r *referenceMultiGetRequest) add(id interface{}, handler referenceResponseHandler) {
	r.ids = append(r.ids, id)
	r.handlers = append(r.handlers, handler)
}

func (r *referenceMultiGetRequest) execute(ctx context.Context, rsc Resource) error {
	subRsc, err := rsc.SubResource(ctx, r.resourcePath)
	if err != nil {
		return err
	}
	payloads, err := subRsc.MultiGet(ctx, r.ids)
	if err != nil {
		return err
	}
	if len(payloads) != len(r.ids) {
		return errors.New("invalid number of items returned by MultiGet")
	}
	validator := subRsc.Validator()
	payloadsWrapper := make([]map[string]interface{}, 1)
	for i, p := range payloads {
		handler := r.handlers[i]
		if p == nil {
			if err := handler(payloadsWrapper[:0], validator); err != nil {
				return err
			}
		}
		payloadsWrapper[0] = p
		if err := handler(payloadsWrapper, validator); err != nil {
			return err
		}
	}
	return nil
}
