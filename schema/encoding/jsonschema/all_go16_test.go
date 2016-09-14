// +build !go1.7

package jsonschema_test

import "testing"

// run is the implementation of Run for Go < 1.7. It logs the testCase name, and blocks until the test has completed.
func (tc *encoderTestCase) run(t *testing.T) {
	t.Logf("--- RUN SUB-TEST %s", tc.name)
	tc.test(t)
}
