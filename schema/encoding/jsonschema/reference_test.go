package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestReferenceValidatorEncode(t *testing.T) {
	testCase := encoderTestCase{
		name: ``,
		schema: schema.Schema{
			Fields: schema.Fields{
				"r": {
					Validator: &schema.Reference{Path: "somewhere"},
				},
			},
		},
		customValidate: fieldValidator("r", `{}`),
	}
	testCase.Run(t)
}
