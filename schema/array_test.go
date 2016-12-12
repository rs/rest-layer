// +build go1.7

package schema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestArrayValidatorCompile(t *testing.T) {
	testCases := []compilerTestCase{
		{
			Name:     "ValuesValidator=&String{}",
			Compiler: &schema.Array{ValuesValidator: &schema.String{}},
		},
		{
			Name:     "ValuesValidator=&String{Regexp:invalid}",
			Compiler: &schema.Array{ValuesValidator: &schema.String{Regexp: "[invalid re"}},
			Error:    "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`",
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}

func TestArrayValidator(t *testing.T) {
	testCases := []fieldValidatorTestCase{
		{
			Name:      `ValuesValidator=nil,Validate([]interface{}{true,"value"})`,
			Validator: &schema.Array{},
			Input:     []interface{}{true, "value"},
			Expect:    []interface{}{true, "value"},
		},
		{
			Name:      `ValuesValidator=&schema.Bool{},Validate([]interface{}{true,false})`,
			Validator: &schema.Array{ValuesValidator: &schema.Bool{}},
			Input:     []interface{}{true, false},
			Expect:    []interface{}{true, false},
		},
		{
			Name:      `ValuesValidator=&schema.Bool{},Validate([]interface{}{true,"value"})`,
			Validator: &schema.Array{ValuesValidator: &schema.Bool{}},
			Input:     []interface{}{true, "value"},
			Error:     "invalid value at #2: not a Boolean",
		},
		{
			Name:      `ValuesValidator=&String{},Validate("value")`,
			Validator: &schema.Array{ValuesValidator: &schema.String{}},
			Input:     "value",
			Error:     "not an array",
		},
		{
			Name:      `MinLen=2,Validate([]interface{}{true,false})`,
			Validator: &schema.Array{ValuesValidator: &schema.Bool{}, MinLen: 2},
			Input:     []interface{}{true, false},
			Expect:    []interface{}{true, false},
		},
		{
			Name:      `MinLen=3,Validate([]interface{}{true,false})`,
			Validator: &schema.Array{ValuesValidator: &schema.Bool{}, MinLen: 3},
			Input:     []interface{}{true, false},
			Error:     "has fewer items than 3",
		},
		{
			Name:      `MaxLen=2,Validate([]interface{}{true,false})`,
			Validator: &schema.Array{ValuesValidator: &schema.Bool{}, MaxLen: 2},
			Input:     []interface{}{true, false},
			Expect:    []interface{}{true, false},
		},
		{
			Name:      `MaxLen=1,Validate([]interface{}{true,false})`,
			Validator: &schema.Array{ValuesValidator: &schema.Bool{}, MaxLen: 1},
			Input:     []interface{}{true, false},
			Error:     "has more items than 1",
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
