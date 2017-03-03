package rest_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/stretchr/testify/assert"
)

// requestTest is a reusable type for testing POST, PUT, PATCH or GET requests. Best used in a map, E.g.:
//     go
type requestTest struct {
	Init           func() *requestTestVars
	NewRequest     func() (*http.Request, error)
	ResponseCode   int
	ResponseHeader http.Header // Only checks provided headers, not that all headers are equal.
	ResponseBody   string
	ExtraTest      func(*testing.T, *requestTestVars)
}

// requestTestVars provodes test runtime variables.
type requestTestVars struct {
	Index   resource.Index             // required
	Storers map[string]resource.Storer // optional: may be used by ExtraTest function
}

// Test runs tt in parallel mode. It can be passed as a second parameter to
// Run(name, f) for the *testing.T type.
func (tt *requestTest) Test(t *testing.T) {
	t.Parallel()
	vars := tt.Init()
	h, err := rest.NewHandler(vars.Index)
	if !assert.NoError(t, err, "rest.NewHandler(vars.Index)") {
		return
	}
	r, err := tt.NewRequest()
	if !assert.NoError(t, err, "tt.NewRequest()") {
		return
	}
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	assert.Equal(t, tt.ResponseCode, w.Code, "h.ServeHTTP(w, r); w.Code")
	headers := w.Header()
	for k, v := range tt.ResponseHeader {
		assert.Equal(t, v, headers[k], "h.ServeHTTP(w, r); w.Header()[%q]", k)
	}
	b, _ := ioutil.ReadAll(w.Body)
	assert.JSONEq(t, tt.ResponseBody, string(b), "h.ServeHTTP(w, r); w.Body")

	if tt.ExtraTest != nil {
		tt.ExtraTest(t, vars)
	}
}
