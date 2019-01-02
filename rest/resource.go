package rest

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema/query"
)

type restResource struct {
	*resource.Resource
}

// Find implements query.Resource interface.
func (r restResource) Find(ctx context.Context, query *query.Query) ([]map[string]interface{}, error) {
	itemList, err := r.Resource.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	payloads := make([]map[string]interface{}, 0, len(itemList.Items))
	for _, i := range itemList.Items {
		payloads = append(payloads, i.Payload)
	}
	return payloads, nil
}

// MultiGet implements query.Resource interface.
func (r restResource) MultiGet(ctx context.Context, ids []interface{}) ([]map[string]interface{}, error) {
	items, err := r.Resource.MultiGet(ctx, ids)
	if err != nil {
		return nil, err
	}
	payloads := make([]map[string]interface{}, 0, len(items))
	for _, i := range items {
		var p map[string]interface{}
		if i != nil {
			p = i.Payload
		}
		payloads = append(payloads, p)
	}
	return payloads, nil
}

// SubResource implements query.Resource interface.
func (r restResource) SubResource(ctx context.Context, path string) (query.Resource, error) {
	router, ok := IndexFromContext(ctx)
	if !ok {
		return restResource{}, errors.New("router not available in context")
	}
	rsc, found := router.GetResource(path, r.Resource)
	if !found {
		return restResource{}, fmt.Errorf("invalid resource reference: %s", path)
	}
	return restResource{rsc}, nil
}
