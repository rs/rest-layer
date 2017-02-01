package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

var tinySchema = schema.Schema{
	Fields: schema.Fields{
		"name": {Validator: &schema.String{}},
		"age":  {Validator: &schema.Integer{}},
	},
}

const tinySchemaJSON = `{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": {"type": "string"},
		"age": {"type": "integer"}
	}
}`

func TestObjectValidatorEncode(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: `Schema=nil`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"o": {
						Validator: &schema.Object{},
					},
				},
			},
			expectError: "no schema defined for object",
		},
		{
			name: `Schema=tiny`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"o": {
						Validator: &schema.Object{Schema: &tinySchema},
					},
				},
			},
			customValidate: fieldValidator("o", tinySchemaJSON),
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
