// +build go1.7

package schema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestDictValidatorCompile(t *testing.T) {
	testCases := []compilerTestCase{
		{
			Name: "KeysValidator=&String{},ValuesValidator=&String{}",
			Compiler: &schema.Dict{
				KeysValidator:   &schema.String{},
				ValuesValidator: &schema.String{},
			},
		},
		{
			Name:     "KeysValidator=&String{Regexp:invalid}",
			Compiler: &schema.Dict{KeysValidator: &schema.String{Regexp: "[invalid re"}},
			Error:    "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`",
		},
		{
			Name:     "ValuesValidator=&String{Regexp:invalid}",
			Compiler: &schema.Dict{ValuesValidator: &schema.String{Regexp: "[invalid re"}},
			Error:    "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`",
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}

func TestDictValidator(t *testing.T) {
	testCases := []fieldValidatorTestCase{
		{
			Name:      `KeysValidator=&String{},Validate(map[string]interface{}{"foo":true,"bar":false})`,
			Validator: &schema.Dict{KeysValidator: &schema.String{}},
			Input:     map[string]interface{}{"foo": true, "bar": false},
			Expect:    map[string]interface{}{"foo": true, "bar": false},
		},
		{
			Name:      `KeysValidator=&String{MinLen:3},Validate(map[string]interface{}{"foo":true,"bar":false})`,
			Validator: &schema.Dict{KeysValidator: &schema.String{MinLen: 3}},
			Input:     map[string]interface{}{"foo": true, "bar": false},
			Expect:    map[string]interface{}{"foo": true, "bar": false},
		},
		{
			Name:      `KeysValidator=&String{MinLen:3},Validate(map[string]interface{}{"foo":true,"ba":false})`,
			Validator: &schema.Dict{KeysValidator: &schema.String{MinLen: 3}},
			Input:     map[string]interface{}{"foo": true, "ba": false},
			Error:     "invalid key `ba': is shorter than 3",
		},
		{
			Name:      `ValuesValidator=&Bool{},Validate(map[string]interface{}{"foo":true,"bar":false})`,
			Validator: &schema.Dict{ValuesValidator: &schema.Bool{}},
			Input:     map[string]interface{}{"foo": true, "bar": false},
			Expect:    map[string]interface{}{"foo": true, "bar": false},
		},
		{
			Name:      `ValuesValidator=&Bool{},Validate(map[string]interface{}{"foo":true,"bar":"value"})`,
			Validator: &schema.Dict{ValuesValidator: &schema.Bool{}},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Error:     "invalid value for key `bar': not a Boolean",
		},
		{
			Name:      `ValuesValidator=&String{},Validate("value")`,
			Validator: &schema.Dict{ValuesValidator: &schema.String{}},
			Input:     "value",
			Error:     "not a dict",
		},
		{
			Name:      `MinLen=2,Validate(map[string]interface{}{"foo":true,"bar":false})`,
			Validator: &schema.Dict{MinLen: 2},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Expect:    map[string]interface{}{"foo": true, "bar": "value"},
		},
		{
			Name:      `MinLen=3,Validate(map[string]interface{}{"foo":true,"bar":false})`,
			Validator: &schema.Dict{MinLen: 3},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Error:     "has fewer properties than 3",
		},
		{
			Name:      `MaxLen=2,Validate(map[string]interface{}{"foo":true,"bar":false})`,
			Validator: &schema.Dict{MaxLen: 3},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Expect:    map[string]interface{}{"foo": true, "bar": "value"},
		},
		{
			Name:      `MaxLen=1,Validate(map[string]interface{}{"foo":true,"bar":false})`,
			Validator: &schema.Dict{MaxLen: 1},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Error:     "has more properties than 1",
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
