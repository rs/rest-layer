// +build go1.7

package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestBoolValidatorEncode(t *testing.T) {
	testCase := encoderTestCase{
		name: ``,
		schema: schema.Schema{
			Fields: schema.Fields{
				"b": {
					Validator: &schema.Bool{},
				},
			},
		},
		customValidate: fieldValidator("b", `{"type": "boolean"}`),
	}
	testCase.Run(t)
}
