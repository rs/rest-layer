package rest

import (
	"context"
	"net/http"
	"testing"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestHandlerOptionsItem(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("OPTIONS", "/test/1", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "1",
				Resource: test,
			},
		},
	}
	status, headers, body := itemOptions(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, http.Header{
		"Allow":       []string{"DELETE, GET, HEAD, PATCH, PUT"},
		"Allow-Patch": []string{"application/json"}}, headers)
	assert.Nil(t, body)
}
