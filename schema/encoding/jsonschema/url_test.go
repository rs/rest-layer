package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestURLValidatorEncode(t *testing.T) {
	testCase := encoderTestCase{
		name: ``,
		schema: schema.Schema{
			Fields: schema.Fields{
				"url": {
					Validator: &schema.URL{},
				},
			},
		},
		customValidate: fieldValidator("ip", `{
			"type": "string",
			"format": "uri"
		}`),
	}
	testCase.Run(t)
}
