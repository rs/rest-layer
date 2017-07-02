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

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
	"github.com/stretchr/testify/assert"
)

func TestHandlerPutItem(t *testing.T) {
	i := resource.NewIndex()
	s := mem.NewHandler()
	i.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}, "name": {}}}, s, resource.DefaultConf)
	h, _ := NewHandler(i)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("PUT", "/foo/1", bytes.NewBufferString(`{"name": "test"}`))
	h.ServeHTTP(w, r)
	assert.Equal(t, 201, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"id\":\"1\",\"name\":\"test\"}", string(b))
	lkp := resource.NewLookupWithQuery(query.Query{query.Equal{Field: "id", Value: "1"}})
	l, err := s.Find(context.TODO(), lkp, 0, 1)
	assert.NoError(t, err)
	assert.Len(t, l.Items, 1)
}

func TestHandlerPutItemBadPayload(t *testing.T) {
	r, _ := http.NewRequest("PUT", "/test", bytes.NewBufferString("{invalid json"))
	status, headers, body := itemPut(context.TODO(), r, nil)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusBadRequest, err.Code)
		assert.Equal(t, "Malformed body: invalid character 'i' looking for beginning of object key string", err.Message)
	}
}

func TestHandlerPutItemInvalidLookupFields(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("PUT", "/test/2", bytes.NewBufferString("{}"))
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
	status, headers, body := itemPut(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `fields` parameter: invalid: unknown field", err.Message)
	}
}

func TestHandlerPutItemModify(t *testing.T) {
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
	r, _ := http.NewRequest("PUT", "/test/2", bytes.NewBufferString(`{"id": "2", "foo": "baz"}`))
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
	status, headers, body := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.Item{}) {
		i := body.(*resource.Item)
		assert.Equal(t, "2", i.ID)
		assert.Equal(t, map[string]interface{}{"id": "2", "foo": "baz"}, i.Payload)
	}
}

func TestHandlerPutItemInvalidField(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, mem.NewHandler(), resource.DefaultConf)
	r, _ := http.NewRequest("PUT", "/test/2", bytes.NewBufferString(`{"foo": "baz"}`))
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
	status, headers, body := itemPut(context.TODO(), r, rm)
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

func TestHandlerPutItemCreateModeNotAllowed(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, mem.NewHandler(), resource.Conf{
		AllowedModes: []resource.Mode{resource.Replace},
	})
	r, _ := http.NewRequest("PUT", "/test/2", bytes.NewBufferString(`{}`))
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
	status, headers, body := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusMethodNotAllowed, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusMethodNotAllowed, err.Code)
		assert.Equal(t, "Invalid method", err.Message)
	}
}

func TestHandlerPutItemCreateModeAllowed(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, mem.NewHandler(), resource.Conf{
		AllowedModes: []resource.Mode{resource.Create},
	})
	r, _ := http.NewRequest("PUT", "/test/2", bytes.NewBufferString(`{}`))
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
	status, _, _ := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusCreated, status)
}

func TestHandlerPutItemReplaceModeNotAllowed(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.Conf{
		AllowedModes: []resource.Mode{resource.Create},
	})
	r, _ := http.NewRequest("PUT", "/test/1", bytes.NewBufferString(`{}`))
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
	status, headers, body := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusMethodNotAllowed, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusMethodNotAllowed, err.Code)
		assert.Equal(t, "Invalid method", err.Message)
	}
}

func TestHandlerPutItemReplaceModeAllowed(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.Conf{
		AllowedModes: []resource.Mode{resource.Replace},
	})
	r, _ := http.NewRequest("PUT", "/test/1", bytes.NewBufferString(`{}`))
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
	status, _, _ := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
}

func TestHandlerPutItemReplaceCannotChangeID(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PUT", "/test/1", bytes.NewBufferString(`{"id": "2"}`))
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
	status, headers, body := itemPut(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Cannot change document ID", err.Message)
	}
}

func TestHandlerPutItemReplaceEtagMatch(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", ETag: "a", Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PUT", "/test/1", bytes.NewBufferString(`{"id": "1"}`))
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
	status, _, _ := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
}

func TestHandlerPutItemReplaceEtagDontMatch(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", ETag: "b", Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PUT", "/test/1", bytes.NewBufferString(`{"id": "1"}`))
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
	status, _, _ := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusPreconditionFailed, status)
}

func TestHandlerPutItemReplaceModifiedSinceMatch(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	yesterday := time.Now().Add(-24 * time.Hour)
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Updated: yesterday, Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PUT", "/test/1", bytes.NewBufferString(`{"id": "1"}`))
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
	status, _, _ := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
}

func TestHandlerPutItemReplaceModifiedSinceDontMatch(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Updated: now, Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PUT", "/test/1", bytes.NewBufferString(`{"id": "1"}`))
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
	status, _, _ := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusPreconditionFailed, status)
}

func TestHandlerPutItemReplaceInvalidModifiedSinceDate(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	now := time.Now()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Updated: now, Payload: map[string]interface{}{"id": "1"}},
	})
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PUT", "/test/1", bytes.NewBufferString(`{"id": "1"}`))
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
	status, _, body := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusBadRequest, status)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusBadRequest, err.Code)
		assert.Equal(t, "Invalid If-Unmodified-Since header", err.Message)
	}
}
func TestHandlerPutItemNoStorage(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{Fields: schema.Fields{"id": {}}}, nil, resource.Conf{
		AllowedModes: []resource.Mode{resource.Replace},
	})
	r, _ := http.NewRequest("PUT", "/test/1", bytes.NewBufferString(`{}`))
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
	status, headers, body := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotImplemented, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusNotImplemented, err.Code)
		assert.Equal(t, "No Storage Defined", err.Message)
	}
}

func TestHandlerPutItemChangePathValue(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "2"}},
		{ID: "2", Payload: map[string]interface{}{"id": "2"}},
	})
	parent := index.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	test := parent.Bind("test", "foo", schema.Schema{Fields: schema.Fields{"id": {}, "foo": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PUT", "/foo/2/test/1", bytes.NewBufferString(`{"id": "1", "foo": "3"}`))
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
	status, headers, body := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.Item{}) {
		i := body.(*resource.Item)
		assert.Equal(t, "1", i.ID)
		assert.Equal(t, map[string]interface{}{"id": "1", "foo": "3"}, i.Payload)
	}
}

func TestHandlerPutItemPathValueNotRemoved(t *testing.T) {
	index := resource.NewIndex()
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "2"}},
		{ID: "2", Payload: map[string]interface{}{"id": "2"}},
	})
	parent := index.Bind("foo", schema.Schema{Fields: schema.Fields{"id": {}}}, s, resource.DefaultConf)
	test := parent.Bind("test", "foo", schema.Schema{Fields: schema.Fields{"id": {}, "foo": {}}}, s, resource.DefaultConf)
	r, _ := http.NewRequest("PUT", "/foo/2/test/1", bytes.NewBufferString(`{"id": "1"}`))
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
	status, headers, body := itemPut(context.TODO(), r, rm)
	assert.Equal(t, http.StatusOK, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &resource.Item{}) {
		i := body.(*resource.Item)
		assert.Equal(t, "1", i.ID)
		assert.Equal(t, map[string]interface{}{"id": "1", "foo": "2"}, i.Payload)
	}
}
