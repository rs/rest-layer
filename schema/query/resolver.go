package query

import (
	"context"
	"errors"
	"sync"

	"github.com/rs/rest-layer/schema"
)

type referenceResponseHandler func(payloads []map[string]interface{}, validator schema.Validator, rsc Resource) error

type referenceBatchResolver struct {
	mu       sync.Mutex
	requests []referenceRequest
	rsc      Resource
}

func (rbr *referenceBatchResolver) request(rsc Resource, q *Query, handler referenceResponseHandler) {
	if len(q.Predicate) == 1 {
		if eq, ok := q.Predicate[0].(*Equal); ok && eq.Field == "id" {
			// Make an optimization for query on a single id so we can coalese them into a single request.
			id := eq.Value
			for _, r := range rbr.requests {
				if r, ok := r.(*referenceMultiGetRequest); ok && rsc.Path() == r.rsc.Path() {
					r.add(id, handler)
					return
				}
			}
			// Not found an existing multi get request for this path, create a new one.
			r := &referenceMultiGetRequest{rsc: rsc}
			r.add(id, handler)
			rbr.appendRequest(r)
			return
		}
	}
	rbr.appendRequest(referenceSingleRequest{
		rsc:     rsc,
		query:   q,
		handler: handler,
	})
}

func (rbr *referenceBatchResolver) appendRequest(r referenceRequest) {
	rbr.mu.Lock()
	defer rbr.mu.Unlock()
	rbr.requests = append(rbr.requests, r)
}

func (rbr *referenceBatchResolver) execute(ctx context.Context) error {
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
				if e := r.execute(ctx); e != nil {
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
	execute(ctx context.Context) error
}

type referenceSingleRequest struct {
	rsc     Resource
	query   *Query
	handler referenceResponseHandler
}

func (r referenceSingleRequest) execute(ctx context.Context) error {
	payloads, err := r.rsc.Find(ctx, r.query)
	if err != nil {
		return err
	}
	return r.handler(payloads, r.rsc.Validator(), r.rsc)
}

type referenceMultiGetRequest struct {
	rsc      Resource
	ids      []interface{}
	handlers []referenceResponseHandler
}

func (r *referenceMultiGetRequest) add(id interface{}, handler referenceResponseHandler) {
	r.ids = append(r.ids, id)
	r.handlers = append(r.handlers, handler)
}

func (r *referenceMultiGetRequest) execute(ctx context.Context) error {
	payloads, err := r.rsc.MultiGet(ctx, r.ids)
	if err != nil {
		return err
	}
	if len(payloads) != len(r.ids) {
		return errors.New("invalid number of items returned by MultiGet")
	}
	validator := r.rsc.Validator()
	payloadsWrapper := make([]map[string]interface{}, 1)
	for i, p := range payloads {
		handler := r.handlers[i]
		if p == nil {
			if err := handler(payloadsWrapper[:0], validator, r.rsc); err != nil {
				return err
			}
		}
		payloadsWrapper[0] = p
		if err := handler(payloadsWrapper, validator, r.rsc); err != nil {
			return err
		}
	}
	return nil
}
