package schema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestAllOfValidatorCompile(t *testing.T) {
	cases := []referenceCompilerTestCase{
		{
			Name:     "{String}",
			Compiler: &schema.AllOf{&schema.String{}},
		},
		{
			Name:     "{String{Regexp:invalid}}",
			Compiler: &schema.AllOf{&schema.String{Regexp: "[invalid re"}},
			Error:    "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`",
		},
	}
	for i := range cases {
		cases[i].Run(t)
	}
}

func TestAllOfValidator(t *testing.T) {
	cases := []fieldValidatorTestCase{
		{
			Name:      "{String}.Validate(true)",
			Validator: schema.AllOf{&schema.Bool{}, &schema.Bool{}},
			Input:     true,
			Expect:    true,
		},
		{
			Name:      `{Bool, String}.Validate("")`,
			Validator: schema.AllOf{&schema.Bool{}, &schema.String{}},
			Input:     "",
			Error:     "not a Boolean",
		},
		{
			Name:      "{Bool, String}.Validate(true)",
			Validator: schema.AllOf{&schema.Bool{}, &schema.String{}},
			Input:     true,
			Error:     "not a string",
		},
		{
			Name: `{Reference{Path:"foo"},Reference{Path:"bar"}}.Validate(validFooReference)`,
			Validator: schema.AllOf{
				&schema.Reference{Path: "foo"},
				&schema.Reference{Path: "bar"},
			},
			ReferenceChecker: fakeReferenceChecker{
				"foo": {
					IDs:             []interface{}{"foo1"},
					Validator:       &schema.String{},
					SchemaValidator: &schema.Schema{},
				},
				"bar": {
					IDs:             []interface{}{"bar1", "bar2", "bar3"},
					Validator:       &schema.String{},
					SchemaValidator: &schema.Schema{},
				},
			},
			Input: "foo1",
			Error: "not found",
		},
	}
	for i := range cases {
		cases[i].Run(t)
	}
}

func TestAllOfQueryValidator(t *testing.T) {
	cases := []fieldQueryValidatorTestCase{
		{
			Name:      "{String}.Validate(true)",
			Validator: schema.AllOf{&schema.Bool{}, &schema.Bool{}},
			Input:     true,
			Expect:    true,
		},
		{
			Name:      `{Bool, String}.Validate("")`,
			Validator: schema.AllOf{&schema.Bool{}, &schema.String{}},
			Input:     "",
			Error:     "not a Boolean",
		},
		{
			Name:      "{Bool, String}.Validate(true)",
			Validator: schema.AllOf{&schema.Bool{}, &schema.String{}},
			Input:     true,
			Error:     "not a string",
		},
		{
			Name: `{Reference{Path:"foo"},Reference{Path:"bar"}}.Validate(validFooReference)`,
			Validator: schema.AllOf{
				&schema.Reference{Path: "foo"},
				&schema.Reference{Path: "bar"},
			},
			ReferenceChecker: fakeReferenceChecker{
				"foo": {
					IDs:             []interface{}{"foo1"},
					Validator:       &schema.String{},
					SchemaValidator: &schema.Schema{},
				},
				"bar": {
					IDs:             []interface{}{"bar1", "bar2", "bar3"},
					Validator:       &schema.String{},
					SchemaValidator: &schema.Schema{},
				},
			},
			Input: "foo1",
			Error: "not found",
		},
	}
	for i := range cases {
		cases[i].Run(t)
	}
}
