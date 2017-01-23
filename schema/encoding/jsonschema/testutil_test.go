package jsonschema_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/encoding/jsonschema"
	"github.com/stretchr/testify/assert"
)

// dummyValidator implemented the schema.FieldValidator interface, but not jsonschema.Builder interface.
type dummyValidator struct{}

func (v dummyValidator) Validate(value interface{}) (interface{}, error) {
	return value, nil
}

// dummyBuilder implemented the schema.FieldValidator and jsonschema.Builder interfaces.
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

// encoderValidator implementations can be used to validate encoded data.
type encoderValidator func(t *testing.T, result []byte)

// fieldValidator returns an encoderValidator that will compare the JSON of v["properties"][fieldName] only, where v is
// a top-level JSONSchema object.
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

// Reusable fields for testing and benchmarks.
var (
	rfc3339NanoField = schema.Field{
		Description: "UTC start time in RFC3339 format with Nano second support, e.g. 2006-01-02T15:04:05.999999999Z",
		Validator: &schema.Time{
			TimeLayouts: []string{time.RFC3339Nano},
		},
	}
)

// Reusable schemas for testing and benchmarks. Will be compiled as part of initialization.
var (
	simpleSchema = schema.Schema{
		Description: "Student and class",
		Fields: schema.Fields{
			"fullName": {
				Description: "The student name",
				Required:    true,
				Validator: &schema.String{
					MinLen: 1,
					MaxLen: 10,
				},
			},
			"class": {
				Description: "The class name",
				Default:     "Unassigned",
				Validator: &schema.String{
					MaxLen: 10,
				},
			},
		},
	}
	nestedObjectsSchema = schema.Schema{
		Description: "Object with a sub-schema for student",
		Fields: schema.Fields{
			"student": {
				Validator: &schema.Object{
					Schema: &simpleSchema,
				},
			},
		},
	}
	arrayOfObjectsSchema = schema.Schema{
		Description: "Object with array of students",
		Fields: schema.Fields{
			"students": {
				Description: "Array of students",
				Validator: &schema.Array{
					ValuesValidator: &schema.Object{
						Schema: &simpleSchema,
					},
				},
			},
		},
	}
)

func init() {
	Must(simpleSchema.Compile())
	Must(nestedObjectsSchema.Compile())
	Must(arrayOfObjectsSchema.Compile())
}

// JSON serialization of reusable schemas.
const (
	simpleSchemaJSON = `{
		"type": "object",
		"description": "Student and class",
		"additionalProperties": false,
		"properties": {
			"fullName": {
				"type": "string",
				"description": "The student name",
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
		"required": ["fullName"]
	}`
	nestedObjectsSchemaJSON = `{
		"type": "object",
		"description": "Object with a sub-schema for student",
		"additionalProperties": false,
		"properties": {
			"student": ` + simpleSchemaJSON + `
		}
	}`
	arrayOfObjectsSchemaJSON = `{
		"type": "object",
		"description": "Object with array of students",
		"additionalProperties": false,
		"properties": {
			"students": {
				"type": "array",
				"description": "Array of students",
				"items": ` + simpleSchemaJSON + `
			}
		}
	}`
)

// Must panics if err is not nil.
func Must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// Default returns a copy of f with f.Default set to v.
func Default(f schema.Field, v interface{}) schema.Field {
	f.Default = v
	return f
}

// Required returns a copy of f with f.Required set to true.
func Required(f schema.Field) schema.Field {
	f.Required = true
	return f
}

// String returns a schema.Field template for string validation.
func String(min, max int, description string) schema.Field {
	return schema.Field{
		Description: description,
		Default:     "",
		Validator: &schema.String{
			MinLen: min,
			MaxLen: max,
		},
	}
}

// Integer returns a schema.Field template for int validation.
func Integer(min, max int, description string) schema.Field {
	return schema.Field{
		Default:     min,
		Description: description,
		Validator: &schema.Integer{
			Boundaries: &schema.Boundaries{
				Min: float64(min),
				Max: float64(max),
			},
		},
	}
}

// Bool returns a schema.Field template for bool validation.
func Bool(description string) schema.Field {
	return schema.Field{
		Description: description,
		Validator:   &schema.Bool{},
	}
}
