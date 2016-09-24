// +build go1.7

package jsonschema_test

import "testing"

// run is the implementation of Run for Go >= 1.7. It relies on Go 1.7 sub-test, and executes in parallel.
func (tc *encoderTestCase) run(t *testing.T) {
	t.Run(tc.name, func(t *testing.T) {
		t.Parallel()
		tc.test(t)
	})
}
