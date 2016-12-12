package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
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

func wrapWithJSONObject(b *bytes.Buffer) []byte {
	return []byte(fmt.Sprintf("{%s}", b.String()))
}

func TestIntegerValidatorNoBoundaryPanic(t *testing.T) {
	validator := &schema.Integer{}
	// Catch regressions in Integer boundary handling
	assert.NotPanics(t, func() { validatorToJSONSchema(new(bytes.Buffer), validator) })
}

func TestStringValidatorNoBoundaryPanic(t *testing.T) {
	validator := &schema.String{}
	// Catch regressions in Integer boundary handling
	assert.NotPanics(t, func() { validatorToJSONSchema(new(bytes.Buffer), validator) })
}

func TestFloatValidator(t *testing.T) {
	validator := &schema.Float{
		Allowed: []float64{23.5, 98.6},
	}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator))
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

func TestTime(t *testing.T) {
	validator := &schema.Time{}
	b := new(bytes.Buffer)
	assert.NoError(t, validatorToJSONSchema(b, validator))
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
	assert.NoError(t, validatorToJSONSchema(b, validator))
	_, err := isValidJSON(wrapWithJSONObject(b))
	assert.NoError(t, err)

	a := assert.New(t)
	strJSON := string(wrapWithJSONObject(b))
	a.Contains(strJSON, `"type": "boolean"`)
}

func TestErrNotImplemented(t *testing.T) {
	validator := &schema.IP{}
	b := new(bytes.Buffer)
	assert.Equal(t, ErrNotImplemented, validatorToJSONSchema(b, validator))
}
