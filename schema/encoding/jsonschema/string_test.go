package jsonschema_test

import (
	"bytes"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/encoding/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestStringValidatorNoBoundaryPanic(t *testing.T) {
	s := schema.Schema{
		Fields: schema.Fields{
			"s": {
				Validator: &schema.String{},
			},
		},
	}
	assert.NotPanics(t, func() {
		enc := jsonschema.NewEncoder(new(bytes.Buffer))
		enc.Encode(&s)
	})
}

func TestStringValidatorEncode(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: `MinLen=0,MaxLen=0,Allowed=nil,Regexp=""`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"s": {
						Validator: &schema.String{},
					},
				},
			},
			customValidate: fieldValidator("s", `{"type": "string"}`),
		},
		// Ensure backspaces are escaped in regular expressions.
		{
			name: `Regexp="\s+$"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"s": {
						Validator: &schema.String{
							Regexp: `\s+$`,
						},
					},
				},
			},
			customValidate: fieldValidator("s", `{"type": "string", "pattern": "\\s+$"}`),
		},
		{
			name: `Allowed=["one","two"]`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"s": {
						Validator: &schema.String{
							Allowed: []string{"one", "two"},
						},
					},
				},
			},
			customValidate: fieldValidator("s", `{"type": "string", "enum": ["one", "two"]}`),
		},
		{
			name: `MinLen=3,MaxLen=23`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"s": {
						Validator: &schema.String{
							MinLen: 3,
							MaxLen: 23,
						},
					},
				},
			},
			customValidate: fieldValidator("s", `{"type": "string", "minLength": 3, "maxLength": 23}`),
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
