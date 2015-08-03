package rest

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testResponseSender struct {
	call []interface{}
}

func (r *testResponseSender) Send(w http.ResponseWriter, status int, data interface{}) {
	r.call = []interface{}{"Send", w, status, data}
}
func (r *testResponseSender) SendError(w http.ResponseWriter, err error, skipBody bool) {
	r.call = []interface{}{"SendError", w, err, skipBody}
}
func (r *testResponseSender) SendItem(w http.ResponseWriter, status int, i *Item, skipBody bool) {
	r.call = []interface{}{"SendItem", w, status, i, skipBody}
}
func (r *testResponseSender) SendList(w http.ResponseWriter, l *ItemList, skipBody bool) {
	r.call = []interface{}{"SendList", w, l, skipBody}
}

func TestRequestSend(t *testing.T) {
	s := &testResponseSender{}
	r := request{
		s: s,
	}
	r.send(200, "")
	assert.Equal(t, []interface{}{"Send", nil, 200, ""}, s.call)
	err := errors.New("error")
	r.sendError(err)
	assert.Equal(t, []interface{}{"SendError", nil, err, false}, s.call)
	i := &Item{}
	r.sendItem(200, i)
	assert.Equal(t, []interface{}{"SendItem", nil, 200, i, false}, s.call)
	l := &ItemList{}
	r.sendList(l)
	assert.Equal(t, []interface{}{"SendList", nil, l, false}, s.call)

}

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
