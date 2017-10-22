package rest

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestHandlerHead(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
		{ID: "4"},
		{ID: "5"},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, s, resource.DefaultConf)
	r, _ := http.NewRequest("HEAD", "/test", nil)
	rm := &RouteMatch{
		Method: "HEAD",
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{},
	}
	status, headers, body := listHead(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.ItemList{}) {
		l := body.(*resource.ItemList)
		assert.Len(t, l.Items, 0)
		assert.Equal(t, -1, l.Total)
	}

	rm.Params.Set("total", "1")

	status, headers, body = listHead(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.ItemList{}) {
		l := body.(*resource.ItemList)
		assert.Len(t, l.Items, 0)
		assert.Equal(t, 5, l.Total)
	}
}
