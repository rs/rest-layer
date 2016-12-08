package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestHandlerGetListInvalidLookupFields(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{
			"fields": []string{"invalid"},
		},
	}
	status, headers, body := listGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `fields` paramter: invalid: unknown field", err.Message)
	}
}

func TestHandlerGetListInvalidLookupSort(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{
			"sort": []string{"invalid"},
		},
	}
	status, headers, body := listGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `sort` paramter: invalid sort field: invalid", err.Message)
	}
}

func TestHandlerGetListInvalidLookupFilter(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test", nil)
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
	status, headers, body := listGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `filter` parameter: must be valid JSON", err.Message)
	}
}

func TestHandlerGetListInvalidPage(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{
			"page": []string{"invalid"},
		},
	}
	status, headers, body := listGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `page` parameter", err.Message)
	}

	rm.Params.Set("page", "-1")

	status, headers, body = listGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `page` parameter", err.Message)
	}
}

func TestHandlerGetListInvalidLimit(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("GET", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{
			"limit": []string{"invalid"},
		},
	}
	status, headers, body := listGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `limit` parameter", err.Message)
	}

	rm.Params.Set("limit", "-1")

	status, headers, body = listGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `limit` parameter", err.Message)
	}
}

func TestHandlerGetListPageWithNoLimit(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.Conf{})
	r, _ := http.NewRequest("GET", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{
			"page": []string{"2"},
		},
	}
	status, headers, body := listGet(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Cannot use `page' parameter with no `limit' parameter on a resource with no default pagination size", err.Message)
	}
}

func TestHandlerGetListPagination(t *testing.T) {
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
	r, _ := http.NewRequest("GET", "/test", nil)
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{
			"page":  []string{"2"},
			"limit": []string{"2"},
		},
	}
	status, headers, body := listGet(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.ItemList{}) {
		l := body.(*resource.ItemList)
		if assert.Len(t, l.Items, 2) {
			assert.Equal(t, "3", l.Items[0].ID)
			assert.Equal(t, "4", l.Items[1].ID)
		}
		assert.Equal(t, 2, l.Page)
		assert.Equal(t, 5, l.Total)
	}

	rm.Params.Set("page", "3")

	status, headers, body = listGet(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.ItemList{}) {
		l := body.(*resource.ItemList)
		if assert.Len(t, l.Items, 1) {
			assert.Equal(t, "5", l.Items[0].ID)
		}
		assert.Equal(t, 3, l.Page)
		assert.Equal(t, 5, l.Total)
	}
}

func TestHandlerGetListFieldHandlerError(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{{ID: "1", Payload: map[string]interface{}{"foo": "bar"}}})
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
				Resource: test,
			},
		},
		Params: url.Values{
			"fields": []string{`foo(bar="baz")`},
		},
	}
	status, headers, body := listGet(context.TODO(), r, rm)
	assert.Equal(t, 520, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 520, err.Code)
		assert.Equal(t, "foo: error", err.Message)
	}
}
