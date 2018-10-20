package schema_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rs/rest-layer/schema"
)

// hexByteArray implements the FieldSerializer interface.
type hexByteArray struct{}

// Validate is a dummy implemetation of the FieldValidator interface implemented
// to allow inclusion in AnyOf.
func (h hexByteArray) Validate(value interface{}) (interface{}, error) {
	return nil, nil
}

func (h hexByteArray) Serialize(value interface{}) (interface{}, error) {
	switch t := value.(type) {
	case []byte:
		if len(t) == 0 {
			return "", nil
		}
		res := "0x"
		for _, v := range t {
			res += fmt.Sprintf("%x", v)
		}
		return res, nil
	default:
		return nil, errors.New("invalid type")
	}
}

func TestAnyOfCompile(t *testing.T) {
	cases := []referenceCompilerTestCase{
		{
			Name:             "{String}",
			Compiler:         &schema.AnyOf{&schema.String{}},
			ReferenceChecker: fakeReferenceChecker{},
		},
		{
			Name:             "{String{Regexp:invalid}}",
			Compiler:         &schema.AnyOf{&schema.String{Regexp: "[invalid re"}},
			ReferenceChecker: fakeReferenceChecker{},
			Error:            "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`",
		},
		{
			Name:             "{Reference{Path:valid}}",
			Compiler:         &schema.AnyOf{&schema.Reference{Path: "items"}},
			ReferenceChecker: fakeReferenceChecker{"items": {IDs: []interface{}{1, 2, 3}, Validator: &schema.Integer{}}},
		},
		{
			Name:             "{Reference{Path:invalid}}",
			Compiler:         &schema.AnyOf{&schema.Reference{Path: "foobar"}},
			ReferenceChecker: fakeReferenceChecker{"items": {IDs: []interface{}{1, 2, 3}, Validator: &schema.Integer{}}},
			Error:            "can't find resource 'foobar'",
		},
	}
	for i := range cases {
		cases[i].Run(t)
	}
}

func TestAnyOfValidate(t *testing.T) {
	cases := []fieldValidatorTestCase{
		{
			Name:      "{Bool,Bool}.Validate(true)",
			Validator: schema.AnyOf{&schema.Bool{}, &schema.Bool{}},
			Input:     true,
			Expect:    true,
		},
		{
			Name:      `{Bool,Bool}.Validate("")`,
			Validator: schema.AnyOf{&schema.Bool{}, &schema.Bool{}},
			Input:     "",
			Error:     "invalid",
		},
		{
			Name:      "{Bool,String}.Validate(true)",
			Validator: schema.AnyOf{&schema.Bool{}, &schema.String{}},
			Input:     true,
			Expect:    true,
		},
		{
			Name:      `{Bool,String}.Validate("")`,
			Validator: schema.AnyOf{&schema.Bool{}, &schema.String{}},
			Input:     "",
			Expect:    "",
		},
		{
			Name: `{Reference{Path:"foo"},Reference{Path:"bar"}}.Validate(validFooReference)`,
			Validator: schema.AnyOf{
				&schema.Reference{Path: "foo"},
				&schema.Reference{Path: "bar"},
			},
			ReferenceChecker: fakeReferenceChecker{
				"foo": {
					IDs:       []interface{}{"foo1"},
					Validator: &schema.String{},
				},
				"bar": {
					IDs:       []interface{}{"bar1", "bar2", "bar3"},
					Validator: &schema.String{},
				},
			},
			Input:  "foo1",
			Expect: "foo1",
		},
	}
	for i := range cases {
		cases[i].Run(t)
	}
}

func TestAnyOfQueryValidate(t *testing.T) {
	cases := []fieldQueryValidatorTestCase{
		{
			Name:      "{Bool,Bool}.Validate(true)",
			Validator: schema.AnyOf{&schema.Bool{}, &schema.Bool{}},
			Input:     true,
			Expect:    true,
		},
		{
			Name:      `{Bool,Bool}.Validate("")`,
			Validator: schema.AnyOf{&schema.Bool{}, &schema.Bool{}},
			Input:     "",
			Error:     "invalid",
		},
		{
			Name:      "{Bool,String}.Validate(true)",
			Validator: schema.AnyOf{&schema.Bool{}, &schema.String{}},
			Input:     true,
			Expect:    true,
		},
		{
			Name:      `{Bool,String}.Validate("")`,
			Validator: schema.AnyOf{&schema.Bool{}, &schema.String{}},
			Input:     "",
			Expect:    "",
		},
		{
			Name: `{Reference{Path:"foo"},Reference{Path:"bar"}}.Validate(validFooReference)`,
			Validator: schema.AnyOf{
				&schema.Reference{Path: "foo"},
				&schema.Reference{Path: "bar"},
			},
			ReferenceChecker: fakeReferenceChecker{
				"foo": {
					IDs:       []interface{}{"foo1"},
					Validator: &schema.String{},
				},
				"bar": {
					IDs:       []interface{}{"bar1", "bar2", "bar3"},
					Validator: &schema.String{},
				},
			},
			Input:  "foo1",
			Expect: "foo1",
		},
	}
	for i := range cases {
		cases[i].Run(t)
	}
}

func TestAnyOfSerialize(t *testing.T) {
	cases := []fieldSerializerTestCase{
		{
			Name:       "{Bool,Bool}.Serialize(true)",
			Serializer: schema.AnyOf{&schema.Bool{}, &schema.Bool{}},
			Input:      true,
			Expect:     true,
		},
		{
			Name:       `{Bool,IP}.Serialize("1.2.3.4")`,
			Serializer: schema.AnyOf{&schema.Bool{}, &schema.IP{}},
			Input:      "1.2.3.4",
			Expect:     "1.2.3.4",
		},
		{
			Name:       `{Bool,IP{StoreBinary:true}}.Serialize("1.2.3.4")`,
			Serializer: schema.AnyOf{&schema.Bool{}, &schema.IP{StoreBinary: true}},
			Input:      []byte{1, 2, 3, 4},
			Expect:     "1.2.3.4",
		},
		{
			Name:       `{hexByteArray,IP{StoreBinary:true}}.Serialize([]byte{1,2,3,4})`,
			Serializer: schema.AnyOf{&hexByteArray{}, &schema.IP{StoreBinary: true}},
			Input:      []byte{1, 2, 3, 4},
			Expect:     "0x1234",
		},
		{
			Name:       `{IP{StoreBinary:true},hexByteArray}.Serialize([]byte{1,2,3,4})`,
			Serializer: schema.AnyOf{&schema.IP{StoreBinary: true}, &hexByteArray{}},
			Input:      []byte{1, 2, 3, 4},
			Expect:     "1.2.3.4",
		},
		// IP.Serialize() returns an error if the input is not a 4 or 16 byte
		// array, so hexByteArray.Serialize() should run.
		{
			Name:       `{IP{StoreBinary:true},hexByteArray}.Serialize([]byte{1,2,3,4,5})`,
			Serializer: schema.AnyOf{&schema.IP{StoreBinary: true}, &hexByteArray{}},
			Input:      []byte{1, 2, 3, 4, 5},
			Expect:     "0x12345",
		},
	}
	for i := range cases {
		cases[i].Run(t)
	}
}
