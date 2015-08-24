package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
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
	i.Bind("foo", resource.New(schema.Schema{
		"f": schema.Field{
			Validator: schema.String{
				Regexp: "[",
			},
		},
	}, nil, resource.DefaultConf))
	_, err := NewHandler(i)
	assert.EqualError(t, err, "foo: schema compilation error: f: not a schema.Validator pointer")
}

func TestHandlerGetTimeout(t *testing.T) {
	var d time.Duration
	var err error
	h, _ := NewHandler(resource.NewIndex())
	h.RequestTimeout = 10 * time.Second
	d, err = h.getTimeout(&http.Request{URL: &url.URL{}})
	assert.NoError(t, err)
	assert.Equal(t, 10*time.Second, d)
	d, err = h.getTimeout(&http.Request{URL: &url.URL{RawQuery: "timeout=1s"}})
	assert.NoError(t, err)
	assert.Equal(t, 1*time.Second, d)
	_, err = h.getTimeout(&http.Request{URL: &url.URL{RawQuery: "timeout=invalid"}})
	assert.EqualError(t, err, "time: invalid duration invalid")
}

func TestHandlerGetContext(t *testing.T) {
	var c context.Context
	var err *Error
	h, _ := NewHandler(resource.NewIndex())
	w := newRecorder()
	defer w.Close()
	c, err = h.getContext(w, &http.Request{URL: &url.URL{}})
	assert.Nil(t, err)
	_, ok := c.Deadline()
	assert.False(t, ok)
	h.RequestTimeout = 10 * time.Second
	c, err = h.getContext(w, &http.Request{URL: &url.URL{}})
	assert.Nil(t, err)
	_, ok = c.Deadline()
	assert.True(t, ok)
	c, err = h.getContext(w, &http.Request{URL: &url.URL{RawQuery: "timeout=invalid"}})
	assert.Equal(t, &Error{422, "Cannot parse timeout parameter: time: invalid duration invalid", nil}, err)
}

func TestHandlerServeHTTP(t *testing.T) {
	i := resource.NewIndex()
	i.Bind("foo", resource.New(schema.Schema{}, nil, resource.DefaultConf))
	h, _ := NewHandler(i)
	w := newRecorder()
	defer w.Close()
	u, _ := url.ParseRequestURI("/foo")
	h.ServeHTTP(w, &http.Request{Method: "GET", URL: u})
	assert.Equal(t, 501, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"code\":501,\"message\":\"No Storage Defined\"}", string(b))
}

func TestHandlerServeHTTPNotFound(t *testing.T) {
	h, _ := NewHandler(resource.NewIndex())
	w := newRecorder()
	defer w.Close()
	u, _ := url.ParseRequestURI("/")
	h.ServeHTTP(w, &http.Request{Method: "GET", URL: u})
	assert.Equal(t, 404, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"code\":404,\"message\":\"Resource Not Found\"}", string(b))
}

func TestHandlerServeHTTPParentNotFound(t *testing.T) {
	i := resource.NewIndex()
	foo := i.Bind("foo", resource.New(schema.Schema{}, mem.NewHandler(), resource.DefaultConf))
	foo.Bind("bar", "f", resource.New(schema.Schema{"f": schema.Field{}}, nil, resource.DefaultConf))
	h, _ := NewHandler(i)
	w := newRecorder()
	defer w.Close()
	u, _ := url.ParseRequestURI("/foo/1/bar/2")
	h.ServeHTTP(w, &http.Request{Method: "GET", URL: u})
	assert.Equal(t, 404, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"code\":404,\"message\":\"Parent Resource Not Found\"}", string(b))
}

func TestHandlerServeHTTPInvalidTimeout(t *testing.T) {
	h, _ := NewHandler(resource.NewIndex())
	w := newRecorder()
	defer w.Close()
	u, _ := url.ParseRequestURI("/?timeout=invalid")
	h.ServeHTTP(w, &http.Request{Method: "GET", URL: u})
	assert.Equal(t, 422, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"code\":422,\"message\":\"Cannot parse timeout parameter: time: invalid duration invalid\"}", string(b))
}
