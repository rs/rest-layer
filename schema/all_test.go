// +build go1.7

package schema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

type compilerTestCase struct {
	Name     string
	Compiler schema.Compiler
	Error    string
}

func (tc compilerTestCase) Run(t *testing.T) {
	t.Run(tc.Name, func(t *testing.T) {
		t.Parallel()

		err := tc.Compiler.Compile()
		if tc.Error == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, tc.Error)
		}
	})
}

type fieldValidatorTestCase struct {
	Name          string
	Validator     schema.FieldValidator
	Input, Expect interface{}
	Error         string
}

func (tc fieldValidatorTestCase) Run(t *testing.T) {
	t.Run(tc.Name, func(t *testing.T) {
		t.Parallel()

		v, err := tc.Validator.Validate(tc.Input)
		if tc.Error == "" {
			assert.NoError(t, err)
			assert.Equal(t, tc.Expect, v)
		} else {
			assert.EqualError(t, err, tc.Error)
			assert.Nil(t, v)
		}
	})
}
