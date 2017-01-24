package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestPasswordValidatorEncode(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: `MinLen=0,MaxLen=0`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"p": {
						Validator: &schema.Password{},
					},
				},
			},
			customValidate: fieldValidator("p", `{"type": "string", "format": "password"}`),
		},
		{
			name: `MinLen=3,MaxLen=23`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"p": {
						Validator: &schema.Password{
							MinLen: 3,
							MaxLen: 23,
						},
					},
				},
			},
			customValidate: fieldValidator("p", `{"type": "string", "format": "password", "minLength": 3, "maxLength": 23}`),
		},
		{
			name: `MaxLen=23`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"s": {
						Validator: &schema.String{
							MaxLen: 23,
						},
					},
				},
			},
			customValidate: fieldValidator("s", `{"type": "string", "maxLength": 23}`),
		},
		{
			name: `MinLen=3`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"s": {
						Validator: &schema.String{
							MinLen: 3,
						},
					},
				},
			},
			customValidate: fieldValidator("s", `{"type": "string", "minLength": 3}`),
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
