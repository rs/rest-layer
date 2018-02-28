// +build go1.7

package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestArray(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: "Values.Validator=nil",
			schema: schema.Schema{
				Fields: schema.Fields{
					"a": schema.Field{
						Validator: &schema.Array{},
					},
				},
			},
			customValidate: fieldValidator("a", `{"type": "array"}`),
		},
		{
			name: "Values.Validator=&schema.Bool{}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"a": schema.Field{
						Validator: &schema.Array{Values: schema.Field{
							Validator: &schema.Bool{},
						}},
					},
				},
			},
			customValidate: fieldValidator("a", `{"type": "array", "items": {"type": "boolean"}}`),
		},
		{
			// http://json-schema.org/latest/json-schema-validation.html#rfc.section.5.11
			name: "MinLen=42",
			schema: schema.Schema{
				Fields: schema.Fields{
					"a": schema.Field{
						Validator: &schema.Array{MinLen: 42},
					},
				},
			},
			customValidate: fieldValidator("a", `{"type": "array", "minItems": 42}`),
		},
		{
			// http://json-schema.org/latest/json-schema-validation.html#rfc.section.5.10
			name: "MaxLen=42",
			schema: schema.Schema{
				Fields: schema.Fields{
					"a": schema.Field{
						Validator: &schema.Array{MaxLen: 42},
					},
				},
			},
			customValidate: fieldValidator("a", `{"type": "array", "maxItems": 42}`),
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
