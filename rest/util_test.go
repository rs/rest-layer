package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/rs/rest-layer/resource"
	"github.com/stretchr/testify/assert"
)

func TestGetMethodHandler(t *testing.T) {
	assert.NotNil(t, getMethodHandler(true, "OPTIONS"))
	assert.NotNil(t, getMethodHandler(true, "HEAD"))
	assert.NotNil(t, getMethodHandler(true, "GET"))
	assert.Nil(t, getMethodHandler(true, "POST"))
	assert.NotNil(t, getMethodHandler(true, "PUT"))
	assert.NotNil(t, getMethodHandler(true, "PATCH"))
	assert.NotNil(t, getMethodHandler(true, "DELETE"))
	assert.Nil(t, nil, getMethodHandler(true, "OTHER"))

	assert.NotNil(t, getMethodHandler(false, "OPTIONS"))
	assert.NotNil(t, getMethodHandler(false, "HEAD"))
	assert.NotNil(t, getMethodHandler(false, "GET"))
	assert.Nil(t, getMethodHandler(false, "PUT"))
	assert.NotNil(t, getMethodHandler(false, "POST"))
	assert.Nil(t, getMethodHandler(false, "PATCH"))
	assert.NotNil(t, getMethodHandler(false, "DELETE"))
	assert.Nil(t, getMethodHandler(false, "OTHER"))
}

func TestIsMethodAllowed(t *testing.T) {
	c := resource.Conf{AllowedModes: resource.ReadWrite}
	assert.True(t, isMethodAllowed(true, "OPTIONS", c))
	assert.True(t, isMethodAllowed(true, "HEAD", c))
	assert.True(t, isMethodAllowed(true, "GET", c))
	assert.False(t, isMethodAllowed(true, "POST", c))
	assert.True(t, isMethodAllowed(true, "PUT", c))
	assert.True(t, isMethodAllowed(true, "PATCH", c))
	assert.True(t, isMethodAllowed(true, "DELETE", c))
	assert.False(t, isMethodAllowed(true, "OTHER", c))

	c = resource.Conf{}
	assert.True(t, isMethodAllowed(true, "OPTIONS", c))
	assert.False(t, isMethodAllowed(true, "HEAD", c))
	assert.False(t, isMethodAllowed(true, "GET", c))
	assert.False(t, isMethodAllowed(true, "POST", c))
	assert.False(t, isMethodAllowed(true, "PUT", c))
	assert.False(t, isMethodAllowed(true, "PATCH", c))
	assert.False(t, isMethodAllowed(true, "DELETE", c))
	assert.False(t, isMethodAllowed(true, "OTHER", c))

	assert.True(t, isMethodAllowed(true, "PUT", resource.Conf{AllowedModes: []resource.Mode{resource.Create}}))
	assert.True(t, isMethodAllowed(true, "PUT", resource.Conf{AllowedModes: []resource.Mode{resource.Replace}}))

	c = resource.Conf{AllowedModes: resource.ReadWrite}
	assert.True(t, isMethodAllowed(false, "OPTIONS", c))
	assert.True(t, isMethodAllowed(false, "HEAD", c))
	assert.True(t, isMethodAllowed(false, "GET", c))
	assert.True(t, isMethodAllowed(false, "POST", c))
	assert.False(t, isMethodAllowed(false, "PUT", c))
	assert.False(t, isMethodAllowed(false, "PATCH", c))
	assert.True(t, isMethodAllowed(false, "DELETE", c))
	assert.False(t, isMethodAllowed(false, "OTHER", c))

	c = resource.Conf{}
	assert.True(t, isMethodAllowed(false, "OPTIONS", c))
	assert.False(t, isMethodAllowed(false, "HEAD", c))
	assert.False(t, isMethodAllowed(false, "GET", c))
	assert.False(t, isMethodAllowed(false, "POST", c))
	assert.False(t, isMethodAllowed(false, "PUT", c))
	assert.False(t, isMethodAllowed(false, "PATCH", c))
	assert.False(t, isMethodAllowed(false, "DELETE", c))
	assert.False(t, isMethodAllowed(false, "OTHER", c))
}

func TestGetAllowedMethodHandler(t *testing.T) {
	c := resource.Conf{AllowedModes: resource.ReadWrite}
	assert.NotNil(t, getAllowedMethodHandler(true, "OPTIONS", c))
	assert.NotNil(t, getAllowedMethodHandler(true, "HEAD", c))
	assert.NotNil(t, getAllowedMethodHandler(true, "GET", c))
	assert.Nil(t, getAllowedMethodHandler(true, "POST", c))
	assert.NotNil(t, getAllowedMethodHandler(true, "PUT", c))
	assert.NotNil(t, getAllowedMethodHandler(true, "PATCH", c))
	assert.NotNil(t, getAllowedMethodHandler(true, "DELETE", c))
	assert.Nil(t, nil, getAllowedMethodHandler(true, "OTHER", c))

	assert.NotNil(t, getAllowedMethodHandler(false, "OPTIONS", c))
	assert.NotNil(t, getAllowedMethodHandler(false, "HEAD", c))
	assert.NotNil(t, getAllowedMethodHandler(false, "GET", c))
	assert.Nil(t, getAllowedMethodHandler(false, "PUT", c))
	assert.NotNil(t, getAllowedMethodHandler(false, "POST", c))
	assert.Nil(t, getAllowedMethodHandler(false, "PATCH", c))
	assert.NotNil(t, getAllowedMethodHandler(false, "DELETE", c))
	assert.Nil(t, getAllowedMethodHandler(false, "OTHER", c))

	c = resource.Conf{AllowedModes: resource.ReadOnly}
	assert.NotNil(t, getAllowedMethodHandler(true, "OPTIONS", c))
	assert.NotNil(t, getAllowedMethodHandler(true, "HEAD", c))
	assert.NotNil(t, getAllowedMethodHandler(true, "GET", c))
	assert.Nil(t, getAllowedMethodHandler(true, "POST", c))
	assert.Nil(t, getAllowedMethodHandler(true, "PUT", c))
	assert.Nil(t, getAllowedMethodHandler(true, "PATCH", c))
	assert.Nil(t, getAllowedMethodHandler(true, "DELETE", c))
	assert.Nil(t, nil, getAllowedMethodHandler(true, "OTHER", c))

	assert.NotNil(t, getAllowedMethodHandler(false, "OPTIONS", c))
	assert.NotNil(t, getAllowedMethodHandler(false, "HEAD", c))
	assert.NotNil(t, getAllowedMethodHandler(false, "GET", c))
	assert.Nil(t, getAllowedMethodHandler(false, "PUT", c))
	assert.Nil(t, getAllowedMethodHandler(false, "POST", c))
	assert.Nil(t, getAllowedMethodHandler(false, "PATCH", c))
	assert.Nil(t, getAllowedMethodHandler(false, "DELETE", c))
	assert.Nil(t, getAllowedMethodHandler(false, "OTHER", c))
}

func TestSetAllowHeader(t *testing.T) {
	getAllow := func(isItem bool, modes []resource.Mode) http.Header {
		h := http.Header{}
		setAllowHeader(h, isItem, resource.Conf{AllowedModes: modes})
		return h
	}

	assert.Equal(t, http.Header{}, getAllow(true, nil))
	assert.Equal(t, http.Header{
		"Allow-Patch": []string{"application/json"},
		"Allow":       []string{"DELETE, GET, HEAD, PATCH, PUT"}},
		getAllow(true, resource.ReadWrite))
	assert.Equal(t, http.Header{
		"Allow-Patch": []string{"application/json"},
		"Allow":       []string{"DELETE, PATCH, PUT"}},
		getAllow(true, resource.WriteOnly))
	assert.Equal(t, http.Header{"Allow": []string{"GET, HEAD"}}, getAllow(true, resource.ReadOnly))

	assert.Equal(t, http.Header{}, getAllow(false, nil))
	assert.Equal(t, http.Header{"Allow": []string{"DELETE, GET, HEAD, POST"}}, getAllow(false, resource.ReadWrite))
	assert.Equal(t, http.Header{"Allow": []string{"DELETE, POST"}}, getAllow(false, resource.WriteOnly))
	assert.Equal(t, http.Header{"Allow": []string{"GET, HEAD"}}, getAllow(false, resource.ReadOnly))
}

func TestCompareEtag(t *testing.T) {
	assert.True(t, compareEtag(`abc`, `abc`))
	assert.True(t, compareEtag(`"abc"`, `abc`))
	assert.False(t, compareEtag(`'abc'`, `abc`))
	assert.False(t, compareEtag(`"abc`, `abc`))
	assert.False(t, compareEtag(``, `abc`))
	assert.False(t, compareEtag(`"cba"`, `abc`))
}

func TestRequestDecodePayload(t *testing.T) {
	r := &http.Request{
		Body: ioutil.NopCloser(bytes.NewBufferString("{\"foo\":\"bar\"}")),
	}
	var p map[string]interface{}
	err := decodePayload(r, &p)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, p)
}

func TestRequestDecodePayloadContentType(t *testing.T) {
	r := &http.Request{
		Header: map[string][]string{"Content-Type": {"application/json"}},
		Body:   ioutil.NopCloser(bytes.NewBufferString("{\"foo\":\"bar\"}")),
	}
	var p map[string]interface{}
	err := decodePayload(r, &p)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, p)
	r = &http.Request{
		Header: map[string][]string{"Content-Type": {"application/json; charset=utf8"}},
		Body:   ioutil.NopCloser(bytes.NewBufferString("{\"foo\":\"bar\"}")),
	}
	err = decodePayload(r, &p)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, p)
}

func TestRequestDecodePayloadWrongContentType(t *testing.T) {
	r := &http.Request{
		Header: map[string][]string{"Content-Type": {"text/plain"}},
		Body:   ioutil.NopCloser(bytes.NewBufferString("{\"foo\":\"bar\"}")),
	}
	var p map[string]interface{}
	err := decodePayload(r, &p)
	assert.Equal(t, &Error{501, "Invalid Content-Type header: `text/plain' not supported", nil}, err)
}

func TestRequestDecodePayloadInvalidJSON(t *testing.T) {
	r := &http.Request{
		Body: ioutil.NopCloser(bytes.NewBufferString("{\"foo\":\"")),
	}
	var p map[string]interface{}
	err := decodePayload(r, &p)
	assert.Equal(t, &Error{400, "Malformed body: unexpected EOF", nil}, err)
}

func TestRequestCheckIntegrityRequestBadDate(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("If-Unmodified-Since", "invalid date")
	err := checkIntegrityRequest(r, &resource.Item{})
	assert.Equal(t, &Error{400, "Invalid If-Unmodified-Since header", nil}, err)
}

func TestRequestCheckIntegrityRequestNoItem(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("If-Match", "something")
	err := checkIntegrityRequest(r, nil)
	assert.Equal(t, ErrNotFound, err)
}

func TestRequestCheckIntegrityEtagMissmatch(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("If-Match", "foo")
	err := checkIntegrityRequest(r, &resource.Item{ETag: "bar"})
	assert.Equal(t, ErrPreconditionFailed, err)
}

func TestRequestCheckIntegrityEtagMatch(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("If-Match", "foo")
	err := checkIntegrityRequest(r, &resource.Item{ETag: "foo"})
	assert.Nil(t, err)
}

func TestRequestCheckIntegrityModifiedDateMissmatch(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("If-Unmodified-Since", time.Now().Add(-24*time.Hour).Format(time.RFC1123))
	err := checkIntegrityRequest(r, &resource.Item{Updated: time.Now()})
	assert.Equal(t, ErrPreconditionFailed, err)
}

func TestRequestCheckIntegrityModifiedDateMatch(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	now := time.Now()
	r.Header.Set("If-Unmodified-Since", now.Format(time.RFC1123))
	err := checkIntegrityRequest(r, &resource.Item{Updated: now})
	assert.Nil(t, err)
}
