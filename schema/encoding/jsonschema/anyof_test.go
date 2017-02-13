package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestAnyOfValidatorEncode(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: `[]`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"a": {
						Validator: &schema.AnyOf{},
					},
				},
			},
			expectError: "at least one schema must be specified",
		},
		{
			name: `[Integer,String]`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"a": {
						Validator: &schema.AnyOf{
							&schema.Integer{},
							&schema.String{},
						},
					},
				},
			},
			customValidate: fieldValidator("a", `{
				"anyOf": [
					{"type": "integer"},
					{"type": "string"}
				]
			}`),
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
