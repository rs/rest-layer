package jsonschema_test

import (
	"math"
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestAllOfValidatorEncode(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: `[]`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"a": {
						Validator: &schema.AllOf{},
					},
				},
			},
			expectError: "at least one schema must be specified",
		},
		{
			name: `[Integer{Max:3},Integer{Min:0}]`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"a": {
						Validator: &schema.AllOf{
							&schema.Integer{
								Boundaries: &schema.Boundaries{
									Min: math.Inf(-1),
									Max: 3,
								},
							},
							&schema.Integer{
								Boundaries: &schema.Boundaries{
									Min: 0,
									Max: math.Inf(1),
								},
							},
						},
					},
				},
			},
			customValidate: fieldValidator("a", `{
				"allOf": [
					{"type": "integer", "maximum": 3},
					{"type": "integer", "minimum": 0}
				]
			}`),
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
