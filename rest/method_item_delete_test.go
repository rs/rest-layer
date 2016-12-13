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

func TestHandlerDeleteItem(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1"}},
		{ID: "2", Payload: map[string]interface{}{"id": "2"}},
		{ID: "3", Payload: map[string]interface{}{"id": "3"}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, s, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test/2", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "2",
				Resource: test,
			},
		},
	}
	status, headers, body := itemDelete(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNoContent, status)
	assert.Nil(t, headers)
	assert.Nil(t, body)

	l, err := s.Find(context.TODO(), resource.NewLookup(), 0, -1)
	assert.NoError(t, err)
	assert.Len(t, l.Items, 2)
}

func TestHandlerDeleteItemFilterCondition(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "bar"}},
		{ID: "2", Payload: map[string]interface{}{"id": "2", "foo": "bar"}},
		{ID: "3", Payload: map[string]interface{}{"id": "3", "foo": "bar"}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{
		Fields: schema.Fields{"foo": {Filterable: true}},
	}, s, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test/2", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "2",
				Resource: test,
			},
		},
		Params: url.Values{
			"filter": []string{`{"foo": "baz"}`},
		},
	}
	status, headers, body := itemDelete(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusNotFound, err.Code)
		assert.Equal(t, "Not Found", err.Message)
	}

	l, err := s.Find(context.TODO(), resource.NewLookup(), 0, -1)
	assert.NoError(t, err)
	assert.Len(t, l.Items, 3)
}

func TestHandlerDeleteItemEtag(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", ETag: "a", Payload: map[string]interface{}{"id": "1"}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{
		Fields: schema.Fields{"foo": {Filterable: true}},
	}, s, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test/2", nil)
	r.Header.Set("If-Match", "a")
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
	status, headers, body := itemDelete(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNoContent, status)
	assert.Nil(t, headers)
	assert.Nil(t, body)

	l, err := s.Find(context.TODO(), resource.NewLookup(), 0, -1)
	assert.NoError(t, err)
	assert.Len(t, l.Items, 0)
}

func TestHandlerDeleteItemWrongEtag(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", ETag: "b", Payload: map[string]interface{}{"id": "1"}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{
		Fields: schema.Fields{"foo": {Filterable: true}},
	}, s, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test/2", nil)
	r.Header.Set("If-Match", "a")
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
	status, headers, body := itemDelete(context.TODO(), r, rm)
	assert.Equal(t, 412, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 412, err.Code)
		assert.Equal(t, "Precondition Failed", err.Message)
	}

	l, err := s.Find(context.TODO(), resource.NewLookup(), 0, -1)
	assert.NoError(t, err)
	assert.Len(t, l.Items, 1)
}

func TestHandlerDeleteItemInvalidFilter(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test/2", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "2",
				Resource: test,
			},
		},
		Params: url.Values{
			"filter": []string{"invalid"},
		},
	}
	status, headers, body := itemDelete(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `filter` parameter: must be valid JSON", err.Message)
	}
}

func TestHandlerDeleteItemNoStorage(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "2",
				Resource: test,
			},
		},
	}
	status, headers, body := itemDelete(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotImplemented, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusNotImplemented, err.Code)
		assert.Equal(t, "No Storage Defined", err.Message)
	}
}
