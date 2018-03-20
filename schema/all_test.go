package schema_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rs/rest-layer/schema"
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
			if err != nil {
				t.Errorf("Compiler.Compile(%v): unexpected error: %v", tc.ReferenceChecker, err)
			}
		} else {
			if err == nil || err.Error() != tc.Error {
				t.Errorf("Compiler.Compile(%v): expected error: %v, got: %v", tc.ReferenceChecker, tc.Error, err)
			}
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
			if err != nil {
				t.Errorf("Validator.Compile(%v): unexpected error: %v", tc.ReferenceChecker, err)
			}
		}

		v, err := tc.Validator.Validate(tc.Input)
		if tc.Error == "" {
			if err != nil {
				t.Errorf("Validator.Validate(%v): unexpected error: %v", tc.ReferenceChecker, err)
			}
			if !reflect.DeepEqual(v, tc.Expect) {
				t.Errorf("Validator.Validate(%v): expected: %v, got: %v", tc.Input, tc.Expect, v)
			}
		} else {
			if err == nil || err.Error() != tc.Error {
				t.Errorf("Validator.Validate(%v): expected error: %v, got: %v", tc.ReferenceChecker, tc.Error, err)
			}
			if v != nil {
				t.Errorf("Validator.Validate(%v): expected: nil, got: %v", tc.Input, v)
			}
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
			if err != nil {
				t.Errorf("Validator.Compile(%v): unexpected error: %v", tc.ReferenceChecker, err)
			}
		}

		s, err := tc.Serializer.Serialize(tc.Input)
		if tc.Error == "" {
			if err != nil {
				t.Errorf("Serializer.Serialize(%v): unexpected error: %v", tc.ReferenceChecker, err)
			}
			if s != tc.Expect {
				t.Errorf("Serializer.Serialize(%v): expected: %v, got: %v", tc.Input, tc.Expect, s)
			}
		} else {
			if err == nil || err.Error() != tc.Error {
				t.Errorf("Serializer.Serialize(%v): expected error: %v, got: %v", tc.ReferenceChecker, tc.Error, err)
			}
			if s != nil {
				t.Errorf("Serializer.Serialize(%v): expected: nil, got: %v", tc.Input, s)
			}
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
