package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func isValidJSON(payload []byte) (map[string]interface{}, error) {
	b := bytes.NewBuffer(payload)
	decoder := json.NewDecoder(b)
	m := make(map[string]interface{})
	err := decoder.Decode(&m)
	return m, err
}

func copyStringToInterface(src []string) []interface{} {
	dst := make([]interface{}, len(src))
	for i, v := range src {
		dst[i] = v
	}
	return dst
}

func TestReadOnlyField(t *testing.T) {
	stringField := schema.Field{
		ReadOnly:  true,
		Validator: &schema.String{},
	}
	s := &schema.Schema{
		Fields: schema.Fields{
			"name": stringField,
		},
	}
	b := new(bytes.Buffer)
	jse := NewEncoder(b)
	assert.NoError(t, jse.Encode(s))
	_, err := isValidJSON(b.Bytes())
	assert.NoError(t, err)
	a := assert.New(t)
	a.Contains(b.String(), `"readOnly": true`)
	a.Contains(b.String(), `"name":`)
	a.Contains(b.String(), `"type": "string"`)
}

func wrapWithJSONObject(b *bytes.Buffer) []byte {
	return []byte(fmt.Sprintf("{%s}", b.String()))
}

func TestBoundaries(t *testing.T) {
	validator := &schema.Integer{
		Boundaries: &schema.Boundaries{Min: 10, Max: 100},
	}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator, false))
	m, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)
	a := assert.New(t)
	a.Equal(validator.Boundaries.Min, m["minimum"])
	a.Equal(validator.Boundaries.Max, m["maximum"])
}

func TestRegexpEscaping(t *testing.T) {
	validator := &schema.String{
		Regexp: `\s+$`,
	}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator, false))
	_, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)
	a := assert.New(t)
	a.Contains(b.String(), string([]byte{'\\', '\\', 's', '+', '$'}))
}

func TestStringValidator(t *testing.T) {
	validator := &schema.String{
		MinLen: 3,
		MaxLen: 23,
	}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator, false))
	m, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)

	a := assert.New(t)
	strJSON := string(wrapWithJSONObject(b))
	// check string for values
	a.Contains(strJSON, "minLength")
	a.Contains(strJSON, "maxLength")
	a.Contains(strJSON, `"type": "string"`)

	// check decoded JSON for values
	a.Equal("string", m["type"])
	a.Equal(float64(validator.MinLen), m["minLength"])
	a.Equal(float64(validator.MaxLen), m["maxLength"])

}

func TestEmptyStringValidator(t *testing.T) {
	validator := &schema.String{}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator, false))
	m, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)

	a := assert.New(t)
	strJSON := string(wrapWithJSONObject(b))
	// check string for absence values
	a.NotContains(strJSON, "minLength")
	a.NotContains(strJSON, "maxLength")
	// check string for values
	a.Contains(strJSON, `"type": "string"`)

	// check decoded JSON for values
	a.Equal("string", m["type"])
	_, ok := m["minLength"]
	a.False(ok)
	_, ok = m["maxLength"]
	a.False(ok)
}

func TestAllowedStringValidation(t *testing.T) {
	validator := &schema.String{
		Allowed: []string{"one", "two"},
	}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator, false))
	m, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)

	a := assert.New(t)
	strJSON := string(wrapWithJSONObject(b))
	a.NotContains(strJSON, "multipleOf")
	a.Contains(strJSON, `"type": "string"`)
	a.Contains(strJSON, `"enum"`)

	a.Equal("string", m["type"])
	assert.Len(t, m["enum"], 2)
	a.Equal(copyStringToInterface([]string{"one", "two"}), m["enum"])
}

func TestIntegerValidatorNoBoundaryPanic(t *testing.T) {
	validator := &schema.Integer{}
	// Catch regressions in Integer boundary handling
	assert.NotPanics(t, func() { validatorToJSONSchema(new(bytes.Buffer), validator, false) })
}

func TestStringValidatorNoBoundaryPanic(t *testing.T) {
	validator := &schema.String{}
	// Catch regressions in Integer boundary handling
	assert.NotPanics(t, func() { validatorToJSONSchema(new(bytes.Buffer), validator, false) })
}

func TestAllowedIntegerValidator(t *testing.T) {
	validator := &schema.Integer{
		Allowed: []int{10, 50},
	}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator, false))
	m, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)

	a := assert.New(t)
	strJSON := string(wrapWithJSONObject(b))
	a.Contains(strJSON, `"type": "integer"`)
	a.Contains(strJSON, `"enum"`)

	a.Equal("integer", m["type"])
	assert.Len(t, m["enum"], 2)
	values, ok := m["enum"].([]interface{})
	a.True(ok)
	a.Equal(float64(validator.Allowed[0]), values[0])
	a.Equal(float64(validator.Allowed[1]), values[1])
}

func TestFloatValidator(t *testing.T) {
	validator := &schema.Float{
		Allowed: []float64{23.5, 98.6},
	}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator, false))
	m, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)

	a := assert.New(t)
	strJSON := string(wrapWithJSONObject(b))
	a.Contains(strJSON, `"type": "number"`)
	a.Contains(strJSON, `"enum"`)

	a.Equal("number", m["type"])
	assert.Len(t, m["enum"], 2)
	values, ok := m["enum"].([]interface{})
	a.True(ok)
	a.Equal(validator.Allowed[0], values[0])
	a.Equal(validator.Allowed[1], values[1])

}

func TestArray(t *testing.T) {
	validator := &schema.Array{}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator, false))
	_, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)

	a := assert.New(t)
	strJSON := string(wrapWithJSONObject(b))
	a.Contains(strJSON, `"type": "array"`)
}

func TestTime(t *testing.T) {
	validator := &schema.Time{}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator, false))
	_, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)

	a := assert.New(t)
	strJSON := string(wrapWithJSONObject(b))
	a.Contains(strJSON, `"type": "string"`)
	a.Contains(strJSON, `"format": "date-time"`)
}

func TestBoolean(t *testing.T) {
	validator := &schema.Bool{}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator, false))
	_, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)

	a := assert.New(t)
	strJSON := string(wrapWithJSONObject(b))
	a.Contains(strJSON, `"type": "boolean"`)
}

func TestObjectWithNilSchema(t *testing.T) {
	employee := &schema.Schema{
		Description: "employee",
		Fields: schema.Fields{
			"name": schema.Field{
				Validator: &schema.String{},
			},
			"startDate": schema.Field{
				Validator: &schema.Time{},
			},
			"location": schema.Field{
				Validator: &schema.Object{},
			},
		},
	}
	employee.Compile()
	b := new(bytes.Buffer)
	encoder := NewEncoder(b)
	assert.NoError(t, encoder.Encode(employee))

	_, err := isValidJSON(b.Bytes())
	assert.NoError(t, err)
}

// Object when used with arrays need to be wrapped in an object.
// Objects when embedded as a field should not be wrapped.
// This tests the wrapping does not occur when embedded as a
// field by validating the JSON document
func TestSchemaWhenEmbeddedInObjectField(t *testing.T) {
	location := &schema.Schema{
		Description: "location",
		Fields: schema.Fields{
			"country": schema.Field{
				Validator: &schema.String{},
			},
		},
	}
	employee := &schema.Schema{
		Description: "employee",
		Fields: schema.Fields{
			"name": schema.Field{
				Validator: &schema.String{},
			},
			"startDate": schema.Field{
				Validator: &schema.Time{},
			},
			"location": schema.Field{
				Validator: &schema.Object{
					Schema: location,
				},
			},
		},
	}
	employee.Compile()
	b := new(bytes.Buffer)
	encoder := NewEncoder(b)
	assert.NoError(t, encoder.Encode(employee))

	_, err := isValidJSON(b.Bytes())
	assert.NoError(t, err)
}

func TestErrNotImplemented(t *testing.T) {
	validator := &schema.IP{}
	b := new(bytes.Buffer)
	assert.Equal(t, ErrNotImplemented, validatorToJSONSchema(b, validator, false))
}

func TestArrayOfObjectsNilSchema(t *testing.T) {
	s := &schema.Schema{
		Description: "A list of students",
		Fields: schema.Fields{
			"students": schema.Field{
				Validator: &schema.Array{
					ValuesValidator: &schema.Object{},
				},
			},
		},
	}
	b := new(bytes.Buffer)
	encoder := NewEncoder(b)
	assert.NoError(t, encoder.Encode(s))

	_, err := isValidJSON(b.Bytes())
	assert.NoError(t, err)
}

func TestArrayOfObjects(t *testing.T) {
	s := &schema.Schema{
		Description: "A list of students",
		Fields: schema.Fields{
			"students": schema.Field{
				Validator: &schema.Array{
					ValuesValidator: &schema.Object{
						Schema: &schema.Schema{
							Fields: schema.Fields{
								"student": schema.Field{
									Description: "a student",
									Required:    true,
									Default:     "Unknown",
									Validator: &schema.String{
										MinLen: 0,
										MaxLen: 10,
									},
								},
								"class": schema.Field{
									Default: "Unassigned",
									Validator: &schema.String{
										MinLen: 0,
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

	b := new(bytes.Buffer)
	encoder := NewEncoder(b)
	assert.NoError(t, encoder.Encode(s))

	m, err := isValidJSON(b.Bytes())
	assert.NoError(t, err)

	a := assert.New(t)
	a.Equal("object", m["type"])
	a.Equal("A list of students", m["title"])
	p, ok := m["properties"].(map[string]interface{})
	a.True(ok)
	a.NotNil(p["students"])
	students, ok := p["students"].(map[string]interface{})
	a.True(ok)

	a.Equal("array", students["type"])
	items, ok := students["items"].(map[string]interface{})
	a.True(ok)

	a.Equal("object", items["type"])

	ip, ok := items["properties"].(map[string]interface{})
	a.True(ok)

	a.Equal(copyStringToInterface([]string{"student"}), ip["required"])

	student, ok := ip["student"].(map[string]interface{})
	a.True(ok)
	a.Equal("a student", student["description"])
	a.Equal("Unknown", student["default"])

	class, ok := ip["class"].(map[string]interface{})
	a.True(ok)
	a.Equal("Unassigned", class["default"])
}

func TestDefaultEncodingWithStringFieldAndIntegerDefault(t *testing.T) {
	s := &schema.Schema{
		Description: "thing",
		Fields: schema.Fields{
			"item": schema.Field{
				Description: "an item",
				Required:    true,
				Default:     42, // deliberate ERROR we put an integer default on a string field!
				Validator:   &schema.String{},
			},
		},
	}
	b := new(bytes.Buffer)
	encoder := NewEncoder(b)
	assert.NoError(t, encoder.Encode(s))

	m, err := isValidJSON(b.Bytes())
	assert.NoError(t, err)

	a := assert.New(t)

	p, ok := m["properties"].(map[string]interface{})
	a.True(ok)
	item, ok := p["item"].(map[string]interface{})
	a.True(ok)

	// Documenting the expected behavior, even though we know
	// it's the potentially wrong. It should cause a run time error
	// and therefore may not be usable in practice. This test
	// confirms the behavior. Schema does not itself try to
	// enforce any type safety on Default values and it is up to
	// the developer

	// It's allowed by the spec in section 6.2

	a.Equal(float64(42), item["default"])
}
