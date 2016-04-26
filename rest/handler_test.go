package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (w closeNotifyRecorder) CloseNotify() <-chan bool {
	return w.closed
}

func (w closeNotifyRecorder) Close() {
	w.closed <- true
}

func newRecorder() *closeNotifyRecorder {
	return &closeNotifyRecorder{httptest.NewRecorder(), make(chan bool, 1)}
}

func TestNewHandler(t *testing.T) {
	i := resource.NewIndex()
	h, err := NewHandler(i)
	assert.NoError(t, err)
	assert.Equal(t, DefaultResponseSender{}, h.ResponseSender)
	assert.Equal(t, i, h.index)
}

func TestNewHandlerNoCompile(t *testing.T) {
	i := resource.NewIndex()
	i.Bind("foo", schema.Schema{
		Fields: schema.Fields{
			"f": {
				Validator: schema.String{
					Regexp: "[",
				},
			},
		},
	}, nil, resource.DefaultConf)
	_, err := NewHandler(i)
	assert.EqualError(t, err, "foo: schema compilation error: f: not a schema.Validator pointer")
}

func TestHandlerFallbackHandlerResourceNotFound(t *testing.T) {
	i := resource.NewIndex()
	h, _ := NewHandler(i)
	fallbacked := false
	h.FallbackHandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fallbacked = true
		w.Write([]byte("ok"))
	}
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assert.True(t, fallbacked)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "ok", string(b))
}

func TestHandlerFallbackHandlerInvalidMethod(t *testing.T) {
	i := resource.NewIndex()
	i.Bind("test", schema.Schema{}, nil, resource.Conf{AllowedModes: []resource.Mode{}})
	h, _ := NewHandler(i)
	fallbacked := false
	h.FallbackHandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fallbacked = true
		w.Write([]byte("ok"))
	}
	r, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assert.True(t, fallbacked)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "ok", string(b))
}

func TestHandlerFallbackHandlerRouteFound(t *testing.T) {
	i := resource.NewIndex()
	i.Bind("test", schema.Schema{}, mem.NewHandler(), resource.DefaultConf)
	h, _ := NewHandler(i)
	fallbacked := false
	h.FallbackHandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fallbacked = true
	}
	r, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assert.False(t, fallbacked)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "[]", string(b))
}

func TestGetContext(t *testing.T) {
	w := newRecorder()
	defer w.Close()
	c := getContext(w, &http.Request{URL: &url.URL{}})
	_, ok := c.Deadline()
	assert.False(t, ok)
}

func TestHandlerServeHTTPNoStorage(t *testing.T) {
	i := resource.NewIndex()
	i.Bind("foo", schema.Schema{}, nil, resource.DefaultConf)
	h, _ := NewHandler(i)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(w, r)
	assert.Equal(t, 501, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"code\":501,\"message\":\"No Storage Defined\"}", string(b))
}

func TestHandlerServeHTTPNotFound(t *testing.T) {
	h, _ := NewHandler(resource.NewIndex())
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	h.ServeHTTP(w, r)
	assert.Equal(t, 404, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"code\":404,\"message\":\"Resource Not Found\"}", string(b))
}

func TestHandlerServeHTTPParentNotFound(t *testing.T) {
	i := resource.NewIndex()
	foo := i.Bind("foo", schema.Schema{}, mem.NewHandler(), resource.DefaultConf)
	foo.Bind("bar", "f", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, resource.DefaultConf)
	h, _ := NewHandler(i)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/foo/1/bar/2", nil)
	h.ServeHTTP(w, r)
	assert.Equal(t, 404, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"code\":404,\"message\":\"Parent Resource Not Found\"}", string(b))
}

func TestHandlerServeHTTPGetEmtpyResource(t *testing.T) {
	i := resource.NewIndex()
	i.Bind("foo", schema.Schema{}, mem.NewHandler(), resource.DefaultConf)
	h, _ := NewHandler(i)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "[]", string(b))
}

func TestHandlerServeHTTPGetNotFoundItem(t *testing.T) {
	i := resource.NewIndex()
	i.Bind("foo", schema.Schema{}, mem.NewHandler(), resource.DefaultConf)
	h, _ := NewHandler(i)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/foo/1", nil)
	h.ServeHTTP(w, r)
	assert.Equal(t, 404, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"code\":404,\"message\":\"Not Found\"}", string(b))
}
