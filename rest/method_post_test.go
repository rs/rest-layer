package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestHandlerPostList(t *testing.T) {
	i := resource.NewIndex()
	s := mem.NewHandler()
	i.Bind("foo", schema.Schema{Fields: schema.Fields{
		"id":  {OnInit: func(ctx context.Context, v interface{}) interface{} { return "1" }},
		"foo": {},
		"bar": {}}}, s, resource.DefaultConf)
	h, _ := NewHandler(i)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/foo", bytes.NewBufferString(`{"foo": "bar"}`))
	h.ServeHTTP(w, r)
	assert.Equal(t, 201, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, `{"foo":"bar","id":"1"}`, string(b))
	lkp := resource.NewLookupWithQuery(schema.Query{schema.Equal{Field: "id", Value: "1"}})
	l, err := s.Find(context.TODO(), lkp, 1, 1)
	assert.NoError(t, err)
	if assert.Len(t, l.Items, 1) {
		assert.Equal(t, map[string]interface{}{"id": "1", "foo": "bar"}, l.Items[0].Payload)
	}
}

func TestHandlerPostListBadPayload(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("POST", "/test/1", bytes.NewBufferString("{invalid json"))
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
	status, headers, body := listPost(context.TODO(), r, rm)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusBadRequest, err.Code)
		assert.Equal(t, "Malformed body: invalid character 'i' looking for beginning of object key string", err.Message)
	}
}

func TestHandlerPostListInvalidLookupFields(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("POST", "/test/2", bytes.NewBufferString("{}"))
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
			"fields": []string{"invalid"},
		},
	}
	status, headers, body := listPost(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `fields` paramter: invalid: unknown field", err.Message)
	}
}

func TestHandlerPostListDup(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "bar"}},
		{ID: "2", Payload: map[string]interface{}{"id": "2", "foo": "bar"}},
		{ID: "3", Payload: map[string]interface{}{"id": "3", "foo": "bar"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{
		"id":  {},
		"foo": {},
	}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"id": "2", "foo": "baz"}`))
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
	status, _, _ := listPost(context.TODO(), r, rm)
	assert.Equal(t, http.StatusConflict, status)
}

func TestHandlerPostListNew(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{
		"id":  {},
		"foo": {},
	}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"id": "2", "foo": "baz"}`))
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "4",
				Resource: test,
			},
		},
	}
	status, headers, body := listPost(context.TODO(), r, rm)
	assert.Equal(t, http.StatusCreated, status)
	assert.Equal(t, http.Header{"Content-Location": []string{"/test/2"}}, headers)
	if assert.IsType(t, body, &resource.Item{}) {
		i := body.(*resource.Item)
		assert.Equal(t, "2", i.ID)
		assert.Equal(t, map[string]interface{}{"id": "2", "foo": "baz"}, i.Payload)
	}
}

func TestHandlerPostListInvalidField(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("POST", "/test/2", bytes.NewBufferString(`{"id": "2", "foo": "baz"}`))
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
	status, headers, body := listPost(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Document contains error(s)", err.Message)
		assert.Equal(t, map[string][]interface{}{
			"foo": []interface{}{"invalid field"}}, err.Issues)
	}
}

func TestHandlerPostListNoStorage(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, nil, resource.Conf{
		AllowedModes: []resource.Mode{resource.Replace},
	})
	r, _ := http.NewRequest("POST", "/test/1", bytes.NewBufferString(`{}`))
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
	status, headers, body := listPost(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotImplemented, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusNotImplemented, err.Code)
		assert.Equal(t, "No Storage Defined", err.Message)
	}
}
