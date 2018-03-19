package schema_test

import (
	"errors"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

type referenceCompilerTestCase struct {
	Name             string
	Compiler         schema.Compiler
	ReferenceChecker schema.ReferenceChecker
	Error            string
}

func (tc referenceCompilerTestCase) Run(t *testing.T) {
	t.Run(tc.Name, func(t *testing.T) {
		t.Parallel()

		err := tc.Compiler.Compile(tc.ReferenceChecker)
		if tc.Error == "" {
			assert.NoError(t, err, "Compiler.Compile(%v)", tc.ReferenceChecker)
		} else {
			assert.EqualError(t, err, tc.Error, "Compiler.Compile(%v)", tc.ReferenceChecker)
		}
	})
}

type fieldValidatorTestCase struct {
	Name             string
	Validator        schema.FieldValidator
	ReferenceChecker schema.ReferenceChecker
	Input, Expect    interface{}
	Error            string
}

func (tc fieldValidatorTestCase) Run(t *testing.T) {
	t.Run(tc.Name, func(t *testing.T) {
		t.Parallel()

		if cmp, ok := tc.Validator.(schema.Compiler); ok {
			err := cmp.Compile(tc.ReferenceChecker)
			assert.NoError(t, err, "Validator.Compile(%v)", tc.ReferenceChecker)
		}

		v, err := tc.Validator.Validate(tc.Input)
		if tc.Error == "" {
			assert.NoError(t, err, "Validator.Validate(%v)", tc.Input)
			assert.Equal(t, tc.Expect, v, "Validator.Validate(%v)", tc.Input)
		} else {
			assert.EqualError(t, err, tc.Error, "Validator.Validate(%v)", tc.Input)
			assert.Nil(t, v, "Validator.Validate(%v)", tc.Input)
		}
	})
}

type fieldSerializerTestCase struct {
	Name             string
	Serializer       schema.FieldSerializer
	ReferenceChecker schema.ReferenceChecker
	Input, Expect    interface{}
	Error            string
}

func (tc fieldSerializerTestCase) Run(t *testing.T) {
	t.Run(tc.Name, func(t *testing.T) {
		t.Parallel()

		if cmp, ok := tc.Serializer.(schema.Compiler); ok {
			err := cmp.Compile(tc.ReferenceChecker)
			assert.NoError(t, err, "Validator.Compile(%v)", tc.ReferenceChecker)
		}

		s, err := tc.Serializer.Serialize(tc.Input)
		if tc.Error == "" {
			assert.NoError(t, err, "Serializer.Serialize(%v)", tc.Input)
			assert.Equal(t, tc.Expect, s, "Serializer.Serialize(%v)", tc.Input)
		} else {
			assert.EqualError(t, err, tc.Error, "Serializer.Serialize(%v)", tc.Input)
			assert.Nil(t, s, "Serializer.Serialize(%v)", tc.Input)
		}
	})
}

type fakeReferenceChecker map[string]struct {
	IDs       []interface{}
	Validator schema.FieldValidator
}

func (rc fakeReferenceChecker) Compile() error {
	for name := range rc {
		if rc[name].Validator == nil {
			continue
		}
		if cmp, ok := rc[name].Validator.(schema.Compiler); ok {
			if err := cmp.Compile(rc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (rc fakeReferenceChecker) ReferenceChecker(path string) schema.FieldValidator {
	rsc, ok := rc[path]
	if !ok {
		return nil
	}
	return schema.FieldValidatorFunc(func(value interface{}) (interface{}, error) {
		var id interface{}
		var err error

		// Sanitize ID from input value.
		if rsc.Validator != nil {
			id, err = rsc.Validator.Validate(value)
			if err != nil {
				return nil, err
			}
		} else {
			id = value
		}
		// Check that the ID exists.
		for _, rscID := range rsc.IDs {
			if id == rscID {
				return id, nil
			}
		}
		return nil, errors.New("not found")
	})
}
