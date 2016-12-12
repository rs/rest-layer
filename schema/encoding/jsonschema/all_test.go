package jsonschema_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/encoding/jsonschema"
	"github.com/stretchr/testify/assert"
)

// encoderTestCase is used to test the Encoder.Encode() function.
type encoderTestCase struct {
	name           string
	schema         schema.Schema
	expect         string
	customValidate encoderValidator
}

// Run runs the testCase according to your Go version. For Go >= 1.7, test cases are run in parallel using the Go 1.7
// sub-test feature. For older versions of Go, the testCase name is simply logged before tests are run sequentially.
func (tc *encoderTestCase) Run(t *testing.T) {
	tc.run(t)
}

// test performs the actual test and is used by both implementations of run.
func (tc *encoderTestCase) test(t *testing.T) {
	b := new(bytes.Buffer)
	enc := jsonschema.NewEncoder(b)
	assert.NoError(t, enc.Encode(&tc.schema))

	if tc.customValidate == nil {
		assert.JSONEq(t, tc.expect, b.String())
	} else {
		tc.customValidate(t, b.Bytes())
	}
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
		// readOnly is a custom extension to JSON Schema, also defined by the Swagger 2.0 Schema Object
		// specification. See http://swagger.io/specification/#schemaObject.
		{
			name: "ReadOnly=true",
			schema: schema.Schema{
				Fields: schema.Fields{
					"name": schema.Field{
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
					"name": schema.Field{
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
					"name": schema.Field{
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
									"name": schema.Field{
										Validator: &schema.String{},
									},
									"age": schema.Field{
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
					"foo": schema.Field{
						Validator: &schema.Bool{},
					},
					"bar": schema.Field{
						Validator: &schema.Bool{},
					},
					"baz": schema.Field{
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
					"foo": schema.Field{
						Validator: &schema.Bool{},
					},
					"bar": schema.Field{
						Validator: &schema.Bool{},
					},
					"baz": schema.Field{
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
					"name": schema.Field{
						Required:  true,
						Validator: &schema.String{},
					},
					"age": schema.Field{
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
					"name": schema.Field{
						Required:  true,
						Validator: &schema.String{},
					},
					"age": schema.Field{
						Required:  true,
						Validator: &schema.Integer{},
					},
					"class": schema.Field{
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
		// Documenting the current behavior, but unsure what's the best approach. schema.Object with no schema
		// appears to be invalid and will cause an error if used.
		{
			name: "Validator=Array,ValuesValidator=Object{Schema:nil}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"students": schema.Field{
						Validator: &schema.Array{
							ValuesValidator: &schema.Object{},
						},
					},
				},
			},
			expect: `{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"students": {
						"type": "array",
						"items": {}
					}
				}
			}`,
		},
		{
			name: "Validator=Array,ValuesValidator=Object{Schema:Student}",
			schema: schema.Schema{
				Description: "Object with array of students",
				Fields: schema.Fields{
					"students": schema.Field{
						Description: "Array of students",
						Validator: &schema.Array{
							ValuesValidator: &schema.Object{
								Schema: &schema.Schema{
									Description: "Student and class",
									Fields: schema.Fields{
										"student": schema.Field{
											Description: "The student name",
											Required:    true,
											Default:     "Unknown",
											Validator: &schema.String{
												MinLen: 1,
												MaxLen: 10,
											},
										},
										"class": schema.Field{
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
			},
			expect: `{
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
			}`,
		},
		{
			name: `Validator=Object,Fields["location"].Validator=Object`,
			schema: schema.Schema{
				Fields: schema.Fields{
					"location": schema.Field{
						Validator: &schema.Object{
							Schema: &schema.Schema{
								Fields: schema.Fields{
									"country": schema.Field{
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
					"location": schema.Field{
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
