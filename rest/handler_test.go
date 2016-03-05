package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
			"f": schema.Field{
				Validator: schema.String{
					Regexp: "[",
				},
			},
		},
	}, nil, resource.DefaultConf)
	_, err := NewHandler(i)
	assert.EqualError(t, err, "foo: schema compilation error: f: not a schema.Validator pointer")
}

func TestGetContext(t *testing.T) {
	w := newRecorder()
	defer w.Close()
	c := getContext(w, &http.Request{URL: &url.URL{}})
	_, ok := c.Deadline()
	assert.False(t, ok)
}

func TestHandlerServeHTTP(t *testing.T) {
	i := resource.NewIndex()
	i.Bind("foo", schema.Schema{}, nil, resource.DefaultConf)
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
	foo := i.Bind("foo", schema.Schema{}, mem.NewHandler(), resource.DefaultConf)
	foo.Bind("bar", "f", schema.Schema{Fields: schema.Fields{"f": schema.Field{}}}, nil, resource.DefaultConf)
	h, _ := NewHandler(i)
	w := newRecorder()
	defer w.Close()
	u, _ := url.ParseRequestURI("/foo/1/bar/2")
	h.ServeHTTP(w, &http.Request{Method: "GET", URL: u})
	assert.Equal(t, 404, w.Code)
	b, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, "{\"code\":404,\"message\":\"Parent Resource Not Found\"}", string(b))
}

func TestRouteHandler(t *testing.T) {

}
