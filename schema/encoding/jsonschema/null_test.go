package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestNullValidatorEncode(t *testing.T) {
	testCase := encoderTestCase{
		name: ``,
		schema: schema.Schema{
			Fields: schema.Fields{
				"n": {
					Validator: &schema.Null{},
				},
			},
		},
		customValidate: fieldValidator("n", `{"type": "null"}`),
	}
	testCase.Run(t)
}
