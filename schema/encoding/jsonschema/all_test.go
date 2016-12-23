// +build go1.7

package jsonschema_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/encoding/jsonschema"
	"github.com/stretchr/testify/assert"
)

// studentSchema serves as a complex nested schema example.
var studentSchema = schema.Schema{
	Description: "Object with array of students",
	Fields: schema.Fields{
		"students": {
			Description: "Array of students",
			Validator: &schema.Array{
				ValuesValidator: &schema.Object{
					Schema: &schema.Schema{
						Description: "Student and class",
						Fields: schema.Fields{
							"student": {
								Description: "The student name",
								Required:    true,
								Default:     "Unknown",
								Validator: &schema.String{
									MinLen: 1,
									MaxLen: 10,
								},
							},
							"class": {
								Description: "The class name",
								Default:     "Unassigned",
								Validator: &schema.String{
									MinLen: 0, // Default value.
									MaxLen: 10,
								},
							},
						},
					},
				},
			},
		},
	},
}

// studentSchemaJSON contains the expected JSON serialization of studentSchema.
const studentSchemaJSON = `{
	"type": "object",
	"description": "Object with array of students",
	"additionalProperties": false,
	"properties": {
		"students": {
			"type": "array",
			"description": "Array of students",
			"items": {
				"type": "object",
				"description": "Student and class",
				"additionalProperties": false,
				"properties": {
					"student": {
						"type": "string",
						"description": "The student name",
						"default": "Unknown",
						"minLength": 1,
						"maxLength": 10
					},
					"class": {
						"type": "string",
						"description": "The class name",
						"default": "Unassigned",
						"maxLength": 10
					}
				},
				"required": ["student"]
			}
		}
	}
}`

type dummyValidator struct{}

func (v dummyValidator) Validate(value interface{}) (interface{}, error) {
	return value, nil
}

type dummyBuilder struct {
	dummyValidator
}

func (f dummyBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	return map[string]interface{}{
		"type": "string",
		"enum": []string{"this", "is", "a", "test"},
	}, nil
}

// encoderTestCase is used to test the Encoder.Encode() function.
type encoderTestCase struct {
	name                string
	schema              schema.Schema
	expect, expectError string
	customValidate      encoderValidator
}

func (tc *encoderTestCase) Run(t *testing.T) {
	t.Run(tc.name, func(t *testing.T) {
		t.Parallel()

		b := new(bytes.Buffer)
		enc := jsonschema.NewEncoder(b)

		if tc.expectError == "" {
			assert.NoError(t, enc.Encode(&tc.schema))
			if tc.customValidate == nil {
				assert.JSONEq(t, tc.expect, b.String())
			} else if tc.expect != "" {
				tc.customValidate(t, b.Bytes())
			}
		} else {
			assert.EqualError(t, enc.Encode(&tc.schema), tc.expectError)
		}

	})
}

// encoderValidator can be used to validate encoded data.
type encoderValidator func(t *testing.T, result []byte)

// fieldValidator returns a encoderValidator that will compare the JSON of v["properties"][fieldName] only, where v is a
// top-level JSONSchema object.
func fieldValidator(fieldName, expected string) encoderValidator {
	return func(t *testing.T, result []byte) {
		v := struct {
			Properties map[string]interface{} `json:"properties"`
		}{}
		err := json.Unmarshal(result, &v)
		assert.NoError(t, err, "Input ('%s') needs to be valid JSON", result)
		actual, err := json.Marshal(v.Properties[fieldName])
		assert.NoError(t, err)
		assert.JSONEq(t, expected, string(actual))
	}
}

func TestEncoder(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: "Validator=&dummyValidator{}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"i": {
						Validator: &dummyValidator{},
					},
				},
			},
			expectError: "not implemented",
		},
		{
			name: "Validator=&dummyBuilder{}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"i": {
						Validator: &dummyBuilder{},
					},
				},
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"i": {
						"type": "string",
						"enum": ["this", "is", "a", "test"]
					}
				}
			}`,
		},
		// readOnly is a custom extension to JSON Schema, also defined by the Swagger 2.0 Schema Object
		// specification. See http://swagger.io/specification/#schemaObject.
		{
			name: "ReadOnly=true",
			schema: schema.Schema{
				Fields: schema.Fields{
					"name": {
						ReadOnly:  true,
						Validator: &schema.String{},
					},
				},
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"name": {
						"type": "string",
						"readOnly": true
					}
				}
			}`,
		},
		{
			name: `type(Default)=string`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"name": {
						Validator: &schema.String{},
						Default:   "Luke",
					},
				},
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"name": {
						"type": "string",
						"default": "Luke"
					}
				}
			}`,
		},
		// Currently we do not validate the type for Default.  Note that according to the JSON Schema
		// Specification Section 6.2 (http://json-schema.org/latest/json-schema-validation.html#anchor101), the
		// default value for a schema is not strictly required to apply to the schema.  It's worth noticing that
		// the Swagger 2.0 variant of JSON Schema (http://swagger.io/specification/#schemaObject), does require
		// the default value to match the schema.
		{
			name: `type(Default)=int`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"name": {
						Validator: &schema.String{},
						Default:   24,
					},
				},
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"name": {
						"type": "string",
						"default": 24
					}
				}
			}`,
		},
		{
			name: `type(Default)=map[string]interface{}`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"student": {
						Validator: &schema.Object{
							Schema: &schema.Schema{
								Fields: schema.Fields{
									"name": {
										Validator: &schema.String{},
									},
									"age": {
										Validator: &schema.Integer{},
									},
								},
							},
						},
						Default: map[string]interface{}{
							"name": "Luke",
							"age":  24,
						},
					},
				},
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"student": {
						"type": "object",
						"additionalProperties": false,
						"properties": {
							"name": {
								"type": "string"
							},
							"age": {
								"type": "integer"
							}
						},
						"default": {
							"name": "Luke",
							"age": 24
						}
					}
				}
			}`,
		},
		{
			name: "MinLen=2",
			schema: schema.Schema{
				Fields: schema.Fields{
					"foo": {
						Validator: &schema.Bool{},
					},
					"bar": {
						Validator: &schema.Bool{},
					},
					"baz": {
						Validator: &schema.Bool{},
					},
				},
				MinLen: 2,
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"foo": {"type": "boolean"},
					"bar": {"type": "boolean"},
					"baz": {"type": "boolean"}
				},
				"minProperties": 2
			}`,
		},
		{
			name: "MaxLen=2",
			schema: schema.Schema{
				Fields: schema.Fields{
					"foo": {
						Validator: &schema.Bool{},
					},
					"bar": {
						Validator: &schema.Bool{},
					},
					"baz": {
						Validator: &schema.Bool{},
					},
				},
				MaxLen: 2,
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"foo": {"type": "boolean"},
					"bar": {"type": "boolean"},
					"baz": {"type": "boolean"}
				},
				"maxProperties": 2
			}`,
		},
		{
			name: "Required=true(1/2)",
			schema: schema.Schema{
				Fields: schema.Fields{
					"name": {
						Required:  true,
						Validator: &schema.String{},
					},
					"age": {
						Validator: &schema.Integer{},
					},
				},
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"name": {
						"type": "string"
					},
					"age": {
						"type": "integer"
					}
				},
				"required": ["name"]
			}`,
		},
		{
			name: "Required=true(2/3)",
			schema: schema.Schema{
				Fields: schema.Fields{
					"name": {
						Required:  true,
						Validator: &schema.String{},
					},
					"age": {
						Required:  true,
						Validator: &schema.Integer{},
					},
					"class": {
						Validator: &schema.String{},
					},
				},
			},
			customValidate: func(t *testing.T, result []byte) {
				var expectProperties = map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"age": map[string]interface{}{
						"type": "integer",
					},
					"class": map[string]interface{}{
						"type": "string",
					},
				}

				v := make(map[string]interface{})
				err := json.Unmarshal(result, &v)
				assert.NoError(t, err, "Input ('%s') needs to be valid JSON", result)

				assert.Equal(t, "object", v["type"], `Unexpected "type" value`)
				assert.Equal(t, false, v["additionalProperties"], `Unexpected "additionalProperties" value`)
				assert.Equal(t, expectProperties, v["properties"], `Unexpected "properties" value`)
				assert.Len(t, v["required"], 2, `Unexpected "required" value`)
				assert.IsType(t, []interface{}{}, v["required"], `Unexpected "required" value`)
				assert.Contains(t, v["required"], "name", `Unexpected "required" value`)
				assert.Contains(t, v["required"], "age", `Unexpected "required" value`)
			},
		},

		{
			name: "Validator=Array,ValuesValidator=Object{Schema:nil}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"students": {
						Validator: &schema.Array{
							ValuesValidator: &schema.Object{},
						},
					},
				},
			},
			expectError: "no schema defined for object",
		},
		{
			name:   "Validator=Array,ValuesValidator=Object{Schema:Student}",
			schema: studentSchema,
			expect: studentSchemaJSON,
		},
		{
			name: `Validator=Object,Fields["location"].Validator=Object`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"location": {
						Validator: &schema.Object{
							Schema: &schema.Schema{
								Fields: schema.Fields{
									"country": {
										Validator: &schema.String{},
									},
								},
							},
						},
					},
				},
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"location": {
						"type": "object",
						"additionalProperties": false,
						"properties": {
							"country": {
								"type": "string"
							}
						}
					}
				}
			}`,
		},
		{
			name: `Incorrectly configured field`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"location": {
						Description: "location of your stuff",
					},
				},
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"location": {
						"description": "location of your stuff"
					}
				}
			}`,
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
