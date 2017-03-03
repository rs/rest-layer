// +build go1.7

package schema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestDictCompile(t *testing.T) {
	testCases := []referenceCompilerTestCase{
		{
			Name: "{KeysValidator:String,ValuesValidator:String}",
			Compiler: &schema.Dict{
				KeysValidator:   &schema.String{},
				ValuesValidator: &schema.String{},
			},
			ReferenceChecker: fakeReferenceChecker{},
		},
		{
			Name:             "{KeysValidator:String{Regexp:invalid}}",
			Compiler:         &schema.Dict{KeysValidator: &schema.String{Regexp: "[invalid re"}},
			ReferenceChecker: fakeReferenceChecker{},
			Error:            "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`",
		},
		{
			Name:             "{ValuesValidator:String{Regexp:invalid}}",
			Compiler:         &schema.Dict{ValuesValidator: &schema.String{Regexp: "[invalid re"}},
			ReferenceChecker: fakeReferenceChecker{},
			Error:            "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`",
		},
		{
			Name:             "{ValuesValidator:Reference{Path:valid}}",
			Compiler:         &schema.Dict{ValuesValidator: &schema.Reference{Path: "foo"}},
			ReferenceChecker: fakeReferenceChecker{"foo": {}},
		},
		{
			Name:             "{ValuesValidator:Reference{Path:invalid}}",
			Compiler:         &schema.Dict{ValuesValidator: &schema.Reference{Path: "bar"}},
			ReferenceChecker: fakeReferenceChecker{"foo": {}},
			Error:            "can't find resource 'bar'",
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}

func TestDictValidate(t *testing.T) {
	testCases := []fieldValidatorTestCase{
		{
			Name:      `{KeysValidator:String}.Validate(valid)`,
			Validator: &schema.Dict{KeysValidator: &schema.String{}},
			Input:     map[string]interface{}{"foo": true, "bar": false},
			Expect:    map[string]interface{}{"foo": true, "bar": false},
		},
		{
			Name:      `{KeysValidator:String{MinLen:3}}.Validate(valid)`,
			Validator: &schema.Dict{KeysValidator: &schema.String{MinLen: 3}},
			Input:     map[string]interface{}{"foo": true, "bar": false},
			Expect:    map[string]interface{}{"foo": true, "bar": false},
		},
		{
			Name:      `{KeysValidator:String{MinLen:3}}.Validate(invalid)`,
			Validator: &schema.Dict{KeysValidator: &schema.String{MinLen: 3}},
			Input:     map[string]interface{}{"foo": true, "ba": false},
			Error:     "invalid key `ba': is shorter than 3",
		},
		{
			Name:      `{ValuesValidator:Bool}.Validate(valid)`,
			Validator: &schema.Dict{ValuesValidator: &schema.Bool{}},
			Input:     map[string]interface{}{"foo": true, "bar": false},
			Expect:    map[string]interface{}{"foo": true, "bar": false},
		},
		{
			Name:      `{ValuesValidator:Bool}.Validate({"foo":true,"bar":"value"})`,
			Validator: &schema.Dict{ValuesValidator: &schema.Bool{}},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Error:     "invalid value for key `bar': not a Boolean",
		},
		{
			Name:      `{ValuesValidator:String}.Validate("")`,
			Validator: &schema.Dict{ValuesValidator: &schema.String{}},
			Input:     "",
			Error:     "not a dict",
		},
		{
			Name:      `{MinLen:2}.Validate({"foo":true,"bar":false})`,
			Validator: &schema.Dict{MinLen: 2},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Expect:    map[string]interface{}{"foo": true, "bar": "value"},
		},
		{
			Name:      `{MinLen=3}.Validate({"foo":true,"bar":false})`,
			Validator: &schema.Dict{MinLen: 3},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Error:     "has fewer properties than 3",
		},
		{
			Name:      `{MaxLen=2}.Validate({"foo":true,"bar":false})`,
			Validator: &schema.Dict{MaxLen: 3},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Expect:    map[string]interface{}{"foo": true, "bar": "value"},
		},
		{
			Name:      `{MaxLen=1}.Validate({"foo":true,"bar":false})`,
			Validator: &schema.Dict{MaxLen: 1},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Error:     "has more properties than 1",
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
