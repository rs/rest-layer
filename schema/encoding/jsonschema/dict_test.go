package jsonschema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestDictValidatorEncode(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: `KeysValidator=nil,Values.Validator=nil}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{},
					},
				},
			},
			customValidate: fieldValidator("d", `{"type": "object", "additionalProperties": true}`),
		},
		{
			name: `KeysValidator=Integer{}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.Integer{},
						},
					},
				},
			},
			expectError: "KeysValidator type not supported",
		},
		{
			name: `KeysValidator=String{}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{"type": "object", "additionalProperties": true}`),
		},
		{
			name: `Values.Validator=Integer{}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							Values: schema.Field{
								Validator: &schema.Integer{},
							},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": {
					"type": "integer"
				}
			}`),
		},
		{
			name: `KeysValidator=String{Regexp:"re"}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{Regexp: "re"},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": false,
				"patternProperties": {
					"re": {}
				}
			}`),
		},
		{
			name: `KeysValidator=String{Regexp:"re"},ValuesValidator=Integer{}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{Regexp: "re"},
							Values: schema.Field{
								Validator: &schema.Integer{},
							},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": false,
				"patternProperties": {
					"re": {
						"type": "integer"
					}
				}
			}`),
		},
		{
			name: `KeysValidator=String{Allowed:["match1"]}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{
								Allowed: []string{"match1"},
							},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": false,
				"patternProperties": {
					"^(match1)$": {}
				}
			}`),
		},
		{
			name: `KeysValidator=String{Allowed:["match1","match2"]}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{
								Allowed: []string{"match1", "match2"},
							},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": false,
				"patternProperties": {
					"^(match1|match2)$": {}
				}
			}`),
		},
		{
			name: `KeysValidator=String{Regexp:"tch",Allowed:["match1","match2"]}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{
								Regexp:  "tch",
								Allowed: []string{"match1", "match2"},
							},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": false,
				"allOf": [
					{"patternProperties": {"^(match1|match2)$": {}}},
					{"patternProperties": {"tch": {}}}
				]
			}`),
		},
		{
			name: `KeysValidator=String{Regexp:"tch",Allowed:["match1","match2"]},ValuesValidator=Integer{}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{
								Regexp:  "tch",
								Allowed: []string{"match1", "match2"},
							},
							Values: schema.Field{
								Validator: &schema.Integer{},
							},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": false,
				"allOf": [
					{"patternProperties": {"^(match1|match2)$": {"type": "integer"}}},
					{"patternProperties": {"tch": {"type": "integer"}}}
				]
			}`),
		},
		{
			name: `KeysValidator=String{MinLen:3},ValuesValidator=nil}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{
								MinLen: 3,
							},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": false,
				"patternProperties": {
					"^.{3,}$": {}
				}
			}`),
		},
		{
			name: `KeysValidator=String{MaxLen:4}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{
								MaxLen: 4,
							},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": false,
				"patternProperties": {
					"^.{0,4}$": {}
				}
			}`),
		},
		{
			name: `KeysValidator=String{MinLen:3,MaxLen:4}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{
								MinLen: 3,
								MaxLen: 4,
							},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": false,
				"patternProperties": {
					"^.{3,4}$": {}
				}
			}`),
		},
		{
			name: `KeysValidator=String{MinLen:3,MaxLen:3}"`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"d": {
						Validator: &schema.Dict{
							KeysValidator: &schema.String{
								MinLen: 3,
								MaxLen: 3,
							},
						},
					},
				},
			},
			customValidate: fieldValidator("d", `{
				"type": "object",
				"additionalProperties": false,
				"patternProperties": {
					"^.{3}$": {}
				}
			}`),
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
