// +build go1.7

package jsonschema_test

import (
	"encoding/json"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: "Validator=dummyValidator",
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
			name: "Validator=dummyBuilder",
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
			name: `Validator=String,type(Default)==string`,
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
			name: `Validator=String,type(Default)==int`,
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
			name: `Validator=Object,type(Default)==map[string]interface{}`,
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
			name:   "Schema=simple",
			schema: simpleSchema,
			expect: simpleSchemaJSON,
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
			name:   "Schema=arrayOfObjects",
			schema: arrayOfObjectsSchema,
			expect: arrayOfObjectsSchemaJSON,
		},
		{
			name:   `Schema=nestedObjects`,
			schema: nestedObjectsSchema,
			expect: nestedObjectsSchemaJSON,
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
