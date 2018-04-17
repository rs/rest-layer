package rest

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/rs/rest-layer/resource/testing/mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
	"github.com/stretchr/testify/assert"
)

func TestHandlerPatchItem(t *testing.T) {
	i := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "bar", "bar": "baz"}},
	})
	i.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}, "foo": {}, "bar": {}}}, s, resource.DefaultConf)
	h, _ := NewHandler(i)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PATCH", "/foo/1", bytes.NewBufferString(`{"foo": "baz"}`))
	h.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, `{"bar":"baz","foo":"baz","id":"1"}`, string(b))
	q := &query.Query{
		Predicate: query.Predicate{query.Equal{Field: "id", Value: "1"}},
		Window:    &query.Window{Limit: 1},
	}
	l, err := s.Find(context.TODO(), q)
	assert.NoError(t, err)
	if assert.Len(t, l.Items, 1) {
		assert.Equal(t, map[string]interface{}{"id": "1", "foo": "baz", "bar": "baz"}, l.Items[0].Payload)
	}
}

func TestHandlerPatchItemBadPayload(t *testing.T) {
	r, _ := http.NewRequest("PATCH", "/test/1", bytes.NewBufferString("{invalid json"))
	status, headers, body := itemPatch(context.TODO(), r, nil)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusBadRequest, err.Code)
		assert.Equal(t, "Malformed body: invalid character 'i' looking for beginning of object key string", err.Message)
	}
}

func TestHandlerPatchItemInvalidQueryFields(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("PATCH", "/test/2", bytes.NewBufferString("{}"))
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
	status, headers, body := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `fields` parameter: invalid: unknown field", err.Message)
	}
}

func TestHandlerPatchItemFound(t *testing.T) {
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
	r, _ := http.NewRequest("PATCH", "/test/2", bytes.NewBufferString(`{"id": "2", "foo": "baz"}`))
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
	status, headers, body := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.Item{}) {
		i := body.(*resource.Item)
		assert.Equal(t, "2", i.ID)
		assert.Equal(t, map[string]interface{}{"id": "2", "foo": "baz"}, i.Payload)
	}
}

func TestHandlerPatchItemNotFound(t *testing.T) {
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
	r, _ := http.NewRequest("PATCH", "/test/2", bytes.NewBufferString(`{"id": "2", "foo": "baz"}`))
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
	status, _, _ := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestHandlerPatchItemInvalidField(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PATCH", "/test/2", bytes.NewBufferString(`{"foo": "baz"}`))
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
	status, headers, body := itemPatch(context.TODO(), r, rm)
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

func TestHandlerPatchItemCannotChangeID(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PATCH", "/test/1", bytes.NewBufferString(`{"id": "2"}`))
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
	status, headers, body := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Cannot change document ID", err.Message)
	}
}

func TestHandlerPatchItemReplaceEtagMatch(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", ETag: "a", Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PATCH", "/test/1", bytes.NewBufferString(`{"id": "1"}`))
	r.Header.Set("If-Match", "W/a")
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
	status, _, _ := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
}

func TestHandlerPatchItemReplaceEtagDontMatch(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", ETag: "b", Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PATCH", "/test/1", bytes.NewBufferString(`{"id": "1"}`))
	r.Header.Set("If-Match", "W/a")
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
	status, _, _ := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, http.StatusPreconditionFailed, status)
}

func TestHandlerPatchItemReplaceModifiedSinceMatch(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	yesterday := time.Now().Add(-24 * time.Hour)
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Updated: yesterday, Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PATCH", "/test/1", bytes.NewBufferString(`{"id": "1"}`))
	r.Header.Set("If-Unmodified-Since", yesterday.Format(time.RFC1123))
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
	status, _, _ := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
}

func TestHandlerPatchItemReplaceModifiedSinceDontMatch(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Updated: now, Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PATCH", "/test/1", bytes.NewBufferString(`{"id": "1"}`))
	r.Header.Set("If-Unmodified-Since", yesterday.Format(time.RFC1123))
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
	status, _, _ := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, http.StatusPreconditionFailed, status)
}

func TestHandlerPatchItemReplaceInvalidModifiedSinceDate(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	now := time.Now()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Updated: now, Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PATCH", "/test/1", bytes.NewBufferString(`{"id": "1"}`))
	r.Header.Set("If-Unmodified-Since", "invalid date")
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
	status, _, body := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, http.StatusBadRequest, status)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusBadRequest, err.Code)
		assert.Equal(t, "Invalid If-Unmodified-Since header", err.Message)
	}
}

func TestHandlerPatchItemNoStorage(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, nil, resource.Conf{
		AllowedModes: []resource.Mode{resource.Replace},
	})
	r, _ := http.NewRequest("PATCH", "/test/1", bytes.NewBufferString(`{}`))
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
	status, headers, body := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotImplemented, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusNotImplemented, err.Code)
		assert.Equal(t, "No Storage Defined", err.Message)
	}
}

func TestHandlerPatchItemChangePathValue(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "2"}},
		{ID: "2", Payload: map[string]interface{}{"id": "2"}},
	})
	parent := index.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	test := parent.Bind("test", "foo", schema.Schema{Fields: schema.Fields{"id": {}, "foo": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PATCH", "/foo/2/test/1", bytes.NewBufferString(`{"id": "1", "foo": "3"}`))
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "foo",
				Field:    "foo",
				Value:    "2",
				Resource: parent,
			},
			&ResourcePathComponent{
				Name:     "test",
				Field:    "id",
				Value:    "1",
				Resource: test,
			},
		},
	}
	status, headers, body := itemPatch(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.Item{}) {
		i := body.(*resource.Item)
		assert.Equal(t, "1", i.ID)
		assert.Equal(t, map[string]interface{}{"id": "1", "foo": "3"}, i.Payload)
	}
}
