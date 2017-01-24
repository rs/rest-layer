package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestIPValidatorEncode(t *testing.T) {
	testCase := encoderTestCase{
		name: ``,
		schema: schema.Schema{
			Fields: schema.Fields{
				"ip": {
					Validator: &schema.IP{},
				},
			},
		},
		customValidate: fieldValidator("ip", `{
			"type": "string",
			"oneOf": [
				{"format": "ipv4"},
				{"format": "ipv6"}
			]
		}`),
	}
	testCase.Run(t)
}
