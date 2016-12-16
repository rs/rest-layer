// +build go1.7

package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestTimeValidatorEncode(t *testing.T) {
	testCase := encoderTestCase{
		name: ``,
		schema: schema.Schema{
			Fields: schema.Fields{
				"t": {
					Validator: &schema.Time{},
				},
			},
		},
		customValidate: fieldValidator("t", `{"type": "string", "format": "date-time"}`),
	}
	testCase.Run(t)
}
