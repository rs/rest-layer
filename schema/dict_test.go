// +build go1.7

package schema_test

import (
	"reflect"
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestDictCompile(t *testing.T) {
	testCases := []referenceCompilerTestCase{
		{
			Name: "{KeysValidator:String,ValuesValidator:String}",
			Compiler: &schema.Dict{
				KeysValidator: &schema.String{},
				Values: schema.Field{
					Validator: &schema.String{},
				},
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
			Name:             "{Values.Validator:String{Regexp:invalid}}",
			Compiler:         &schema.Dict{Values: schema.Field{Validator: &schema.String{Regexp: "[invalid re"}}},
			ReferenceChecker: fakeReferenceChecker{},
			Error:            "invalid regexp: error parsing regexp: missing closing ]: `[invalid re`",
		},
		{
			Name:             "{Values.Validator:Reference{Path:valid}}",
			Compiler:         &schema.Dict{Values: schema.Field{Validator: &schema.Reference{Path: "foo"}}},
			ReferenceChecker: fakeReferenceChecker{"foo": {SchemaValidator: &schema.Schema{}}},
		},
		{
			Name:             "{Values.Validator:Reference{Path:invalid}}",
			Compiler:         &schema.Dict{Values: schema.Field{Validator: &schema.Reference{Path: "bar"}}},
			ReferenceChecker: fakeReferenceChecker{"foo": {SchemaValidator: &schema.Schema{}}},
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
			Name:      `{Values.Validator:Bool}.Validate(valid)`,
			Validator: &schema.Dict{Values: schema.Field{Validator: &schema.Bool{}}},
			Input:     map[string]interface{}{"foo": true, "bar": false},
			Expect:    map[string]interface{}{"foo": true, "bar": false},
		},
		{
			Name:      `{Values.Validator:Bool}.Validate({"foo":true,"bar":"value"})`,
			Validator: &schema.Dict{Values: schema.Field{Validator: &schema.Bool{}}},
			Input:     map[string]interface{}{"foo": true, "bar": "value"},
			Error:     "invalid value for key `bar': not a Boolean",
		},
		{
			Name:      `{Values.Validator:String}.Validate("")`,
			Validator: &schema.Dict{Values: schema.Field{Validator: &schema.String{}}},
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

func TestDictGetField(t *testing.T) {
	f := schema.Field{Description: "foobar", Filterable: true}
	t.Run("{KeysValidator=nil}.GetField(valid)", func(t *testing.T) {
		d := schema.Dict{KeysValidator: nil, Values: f}
		if gf := d.GetField("something"); !reflect.DeepEqual(f, *gf) {
			t.Errorf("d.GetField(valid) returned %#v, expected %#v", *gf, f)
		}
	})

	t.Run("{KeysValidator=String}.GetField(valid)", func(t *testing.T) {
		d := schema.Dict{
			KeysValidator: schema.String{Allowed: []string{"foo", "bar"}},
			Values:        f,
		}
		if gf := d.GetField("foo"); !reflect.DeepEqual(f, *gf) {
			t.Errorf("d.GetField(valid) returned %#v, expected %#v", *gf, f)
		}
	})

	t.Run("{KeysValidator=String}.GetField(invalid)", func(t *testing.T) {
		d := schema.Dict{
			KeysValidator: schema.String{Allowed: []string{"foo", "bar"}},
			Values:        f,
		}
		if gf := d.GetField("invalid"); gf != nil {
			t.Errorf("d.GetField(invalid) returned %#v, expected nil", *gf)
		}
	})
}
