package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestDecodePayload(t *testing.T) {
	r := request{
		req: &http.Request{
			Body: ioutil.NopCloser(bytes.NewBufferString("{\"foo\":\"bar\"}")),
		},
	}
	var p map[string]interface{}
	err := r.decodePayload(&p)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, p)
}

func TestRequestDecodePayloadContentType(t *testing.T) {
	r := request{
		req: &http.Request{
			Header: map[string][]string{"Content-Type": []string{"application/json"}},
			Body:   ioutil.NopCloser(bytes.NewBufferString("{\"foo\":\"bar\"}")),
		},
	}
	var p map[string]interface{}
	err := r.decodePayload(&p)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, p)
	r = request{
		req: &http.Request{
			Header: map[string][]string{"Content-Type": []string{"application/json; charset=utf8"}},
			Body:   ioutil.NopCloser(bytes.NewBufferString("{\"foo\":\"bar\"}")),
		},
	}
	err = r.decodePayload(&p)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, p)
}

func TestRequestDecodePayloadWrongContentType(t *testing.T) {
	r := request{
		req: &http.Request{
			Header: map[string][]string{"Content-Type": []string{"text/plain"}},
			Body:   ioutil.NopCloser(bytes.NewBufferString("{\"foo\":\"bar\"}")),
		},
	}
	var p map[string]interface{}
	err := r.decodePayload(&p)
	assert.Equal(t, &Error{501, "Invalid Content-Type header: `text/plain' not supported", nil}, err)
}

func TestRequestDecodePayloadInvalidJSON(t *testing.T) {
	r := request{
		req: &http.Request{
			Body: ioutil.NopCloser(bytes.NewBufferString("{\"foo\":\"")),
		},
	}
	var p map[string]interface{}
	err := r.decodePayload(&p)
	assert.Equal(t, &Error{400, "Malformed body: unexpected EOF", nil}, err)
}

func TestRequestCheckIntegrityRequest(t *testing.T) {
}
