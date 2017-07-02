package rest

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestHandlerDeleteList(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{}},
		{ID: "2", Payload: map[string]interface{}{}},
		{ID: "3", Payload: map[string]interface{}{}},
		{ID: "4", Payload: map[string]interface{}{}},
		{ID: "5", Payload: map[string]interface{}{}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, s, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test", bytes.NewBufferString("{}"))
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
	}
	status, headers, body := listDelete(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNoContent, status)
	assert.Equal(t, http.Header{"X-Total": []string{"5"}}, headers)
	assert.Nil(t, body)

	l, err := s.Find(context.TODO(), resource.NewLookup(), 0, -1)
	assert.NoError(t, err)
	assert.Len(t, l.Items, 0)
}

func TestHandlerDeleteListFilter(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"foo": "bar"}},
		{ID: "2", Payload: map[string]interface{}{"foo": "bar"}},
		{ID: "3", Payload: map[string]interface{}{"foo": "baz"}},
		{ID: "4", Payload: map[string]interface{}{"foo": "baz"}},
		{ID: "5", Payload: map[string]interface{}{}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{
		Fields: schema.Fields{"foo": {Filterable: true}},
	}, s, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test", bytes.NewBufferString("{}"))
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{
			"filter": []string{`{"foo": "bar"}`},
		},
	}
	status, headers, body := listDelete(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNoContent, status)
	assert.Equal(t, http.Header{"X-Total": []string{"2"}}, headers)
	assert.Nil(t, body)

	l, err := s.Find(context.TODO(), resource.NewLookup(), 0, -1)
	assert.NoError(t, err)
	if assert.Len(t, l.Items, 3) {
		assert.Equal(t, "3", l.Items[0].ID)
		assert.Equal(t, "4", l.Items[1].ID)
		assert.Equal(t, "5", l.Items[2].ID)
	}
}

func TestHandlerDeleteListInvalidFilter(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test", bytes.NewBufferString("{}"))
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{
			"filter": []string{"invalid"},
		},
	}
	status, headers, body := listDelete(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `filter` parameter: char 0: expected '{' got 'i'", err.Message)
	}
}

func TestHandlerDeleteListNoStorage(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test", bytes.NewBufferString("{}"))
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
	}
	status, headers, body := listDelete(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotImplemented, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusNotImplemented, err.Code)
		assert.Equal(t, "No Storage Defined", err.Message)
	}
}
