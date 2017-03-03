package schema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

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
