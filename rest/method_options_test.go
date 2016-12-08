package rest

import (
	"context"
	"net/http"
	"testing"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestHandlerOptionsList(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("OPTIONS", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
	}
	status, headers, body := listOptions(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, http.Header{"Allow": []string{"DELETE, GET, HEAD, POST"}}, headers)
	assert.Nil(t, body)
}
