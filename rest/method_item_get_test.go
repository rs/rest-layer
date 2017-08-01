package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestHandlerGetItemInvalidQueryFields(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test/1", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "1",
				Resource: test,
			},
		},
		Params: url.Values{
			"fields": []string{"invalid"},
		},
	}
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `fields` parameter: invalid: unknown field", err.Message)
	}
}

func TestHandlerGetItemInvalidQuerySort(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "1",
				Resource: test,
			},
		},
		Params: url.Values{
			"sort": []string{"invalid"},
		},
	}
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `sort` parameter: invalid sort field: invalid", err.Message)
	}
}

func TestHandlerGetItemInvalidQueryFilter(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "1",
				Resource: test,
			},
		},
		Params: url.Values{
			"filter": []string{"invalid"},
		},
	}
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `filter` parameter: char 0: expected '{' got 'i'", err.Message)
	}
}

func TestHandlerGetItem(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1"}},
		{ID: "2", Payload: map[string]interface{}{"id": "2"}},
		{ID: "3", Payload: map[string]interface{}{"id": "3"}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, s, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test/2", nil)
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
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.Item{}) {
		i := body.(*resource.Item)
		assert.Equal(t, "2", i.ID)
	}
}

func TestHandlerGetItemEtagMatch(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", ETag: "a", Payload: map[string]interface{}{"id": "1"}},
		{ID: "2", ETag: "a", Payload: map[string]interface{}{"id": "2"}},
		{ID: "3", ETag: "a", Payload: map[string]interface{}{"id": "3"}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, s, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test/2", nil)
	r.Header.Set("If-None-Match", "a")
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
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotModified, status)
	assert.Nil(t, headers)
	assert.Nil(t, body)
}

func TestHandlerGetItemEtagDontMatch(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", ETag: "a", Payload: map[string]interface{}{"id": "1"}},
		{ID: "2", ETag: "b", Payload: map[string]interface{}{"id": "2"}},
		{ID: "3", ETag: "a", Payload: map[string]interface{}{"id": "3"}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, s, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test/2", nil)
	r.Header.Set("If-None-Match", "a")
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
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.Item{}) {
		i := body.(*resource.Item)
		assert.Equal(t, "2", i.ID)
	}
}

func TestHandlerGetItemModifiedMatch(t *testing.T) {
	s := mem.NewHandler()
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Updated: yesterday, Payload: map[string]interface{}{"id": "1"}},
		{ID: "2", Updated: yesterday, Payload: map[string]interface{}{"id": "2"}},
		{ID: "3", Updated: yesterday, Payload: map[string]interface{}{"id": "3"}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, s, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test/2", nil)
	r.Header.Set("If-Modified-Since", now.Format(time.RFC1123))
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
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotModified, status)
	assert.Nil(t, headers)
	assert.Nil(t, body)
}

func TestHandlerGetItemModifiedDontMatch(t *testing.T) {
	s := mem.NewHandler()
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Updated: now, Payload: map[string]interface{}{"id": "1"}},
		{ID: "2", Updated: now, Payload: map[string]interface{}{"id": "2"}},
		{ID: "3", Updated: now, Payload: map[string]interface{}{"id": "3"}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, s, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test/2", nil)
	r.Header.Set("If-Modified-Since", yesterday.Format(time.RFC1123))
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
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.Item{}) {
		i := body.(*resource.Item)
		assert.Equal(t, "2", i.ID)
	}
}

func TestHandlerGetItemInvalidIfModifiedSince(t *testing.T) {
	s := mem.NewHandler()
	now := time.Now()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Updated: now, Payload: map[string]interface{}{"id": "1"}},
		{ID: "2", Updated: now, Payload: map[string]interface{}{"id": "2"}},
		{ID: "3", Updated: now, Payload: map[string]interface{}{"id": "3"}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, s, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test/2", nil)
	r.Header.Set("If-Modified-Since", "invalid date")
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
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusBadRequest, err.Code)
		assert.Equal(t, "Invalid If-Modified-Since header", err.Message)
	}
}

func TestHandlerGetItemFieldHandlerError(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "bar"}}})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{
		Fields: schema.Fields{
			"foo": {
				Params: map[string]schema.Param{
					"bar": {},
				},
				Handler: func(ctx context.Context, value interface{}, params map[string]interface{}) (interface{}, error) {
					return nil, errors.New("error")
				},
			},
		},
	}, s, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "1",
				Resource: test,
			},
		},
		Params: url.Values{
			"fields": []string{`foo(bar="baz")`},
		},
	}
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, 520, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 520, err.Code)
		assert.Equal(t, "foo: error", err.Message)
	}
}

func TestHandlerGetItemNoStorage(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test/1", nil)
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
	status, headers, body := itemGet(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotImplemented, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusNotImplemented, err.Code)
		assert.Equal(t, "No Storage Defined", err.Message)
	}
}
