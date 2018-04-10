package rest_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/rest-layer/internal/testutil"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
)

// requestTest is a reusable type for testing POST, PUT, PATCH or GET requests. Best used in a map, E.g.:
//     go
type requestTest struct {
	Init           func() *requestTestVars
	NewRequest     func() (*http.Request, error)
	ResponseCode   int
	ResponseHeader http.Header // Only checks provided headers, not that all headers are equal.
	ResponseBody   string
	ExtraTest      requestCheckerFunc
}

type requestCheckerFunc func(*testing.T, *requestTestVars)

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
	if err != nil {
		t.Errorf("rest.NewHandler failed: %s", err)
		return
	}
	r, err := tt.NewRequest()
	if err != nil || r == nil {
		t.Errorf("tt.NewRequest failed: %s", err)
		return
	}
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	if tt.ResponseCode != w.Code {
		t.Errorf("Expected HTTP response code %d, got %d", tt.ResponseCode, w.Code)
	}
	header := w.Header()
	for k, evs := range tt.ResponseHeader {
		if eCnt, aCnt := len(evs), len(header[k]); eCnt != aCnt {
			t.Errorf("expected HTTP Header %q to have %d items, got %d items", k, eCnt, aCnt)
			continue
		}
		for i, ev := range evs {
			if av := header[k][i]; ev != av {
				t.Errorf("Expected HTTP header[%q][%d] to equal %q, got %q", k, i, ev, av)
			}
		}

	}
	b, _ := ioutil.ReadAll(w.Body)
	if len(tt.ResponseBody) > 0 {
		testutil.JSONEq(t, []byte(tt.ResponseBody), b)
	} else if len(b) > 0 {
		t.Errorf("Expected empty response body, got:\n%s", b)
	}

	if tt.ExtraTest != nil {
		tt.ExtraTest(t, vars)
	}
}
