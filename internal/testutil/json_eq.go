package testutil

import (
	"encoding/json"
	"reflect"
	"testing"
)

// JSONEq logs an error to t if expect and actual does not contain equivalent
// JSON structures, or if either results in a JSON unmarshal error.
// Returns true if no error was logged.
func JSONEq(t testing.TB, expect, actual []byte) bool {
	t.Helper()
	var ei interface{}
	var ai interface{}

	if err := json.Unmarshal(expect, &ei); err != nil {
		t.Errorf("failed to unmarshal JSON from expect:\ninput: %q\nerror: %s", expect, err.Error())
		return false
	}
	if err := json.Unmarshal(actual, &ai); err != nil {
		t.Errorf("failed to unmarshal JSON from actual:\ninput: %q\nerror: %s", actual, err.Error())
		return false
	}

	if !reflect.DeepEqual(ei, ai) {
		// FIXME: a future version should probably log a JSON diff instead.
		t.Errorf("JSON not equal\nexpect: `%s`\nactual: `%s`", expect, actual)
		return false
	}
	return true
}
