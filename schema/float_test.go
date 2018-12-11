package schema_test

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rs/rest-layer/schema"
)

func TestFloatQueryValidator(t *testing.T) {
	cases := []struct {
		name          string
		field         schema.Float
		input, expect interface{}
		err           error
	}{
		{`Float.ValidateQuery(float64)`, schema.Float{}, 1.2, 1.2, nil},
		{`Float.ValidateQuery(int)`, schema.Float{}, 1, nil, errors.New("not a float")},
		{`Float.ValidateQuery(string)`, schema.Float{}, "1.2", nil, errors.New("not a float")},
		{"Float.ValidateQuery(float64)-out of range above", schema.Float{Boundaries: &schema.Boundaries{Min: 0, Max: 2}}, 3.1, 3.1, nil},
		{"Float.ValidateQuery(float64)-in range", schema.Float{Boundaries: &schema.Boundaries{Min: 0, Max: 2}}, 1.1, 1.1, nil},
		{"Float.ValidateQuery(float64)-out of range below", schema.Float{Boundaries: &schema.Boundaries{Min: 2, Max: 10}}, 1.1, 1.1, nil},
	}
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := (tt.field).ValidateQuery(tt.input)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("unexpected error for `%v`\ngot:  %v\nwant: %v", tt.input, err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("invalid output for `%v`:\ngot:  %#v\nwant: %#v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestFloatValidator(t *testing.T) {
	s, err := schema.Float{}.Validate(1.2)
	assert.NoError(t, err)
	assert.Equal(t, 1.2, s)
	s, err = schema.Float{}.Validate(1)
	assert.EqualError(t, err, "not a float")
	assert.Nil(t, s)
	s, err = schema.Float{}.Validate("1.2")
	assert.EqualError(t, err, "not a float")
	assert.Nil(t, s)
	s, err = schema.Float{Boundaries: &schema.Boundaries{Min: 0, Max: 2}}.Validate(3.1)
	assert.EqualError(t, err, "is greater than 2.00")
	assert.Nil(t, s)
	s, err = schema.Float{Boundaries: &schema.Boundaries{Min: 0, Max: 2}}.Validate(1.1)
	assert.NoError(t, err)
	assert.Equal(t, 1.1, s)
	s, err = schema.Float{Boundaries: &schema.Boundaries{Min: 2, Max: 10}}.Validate(1.1)
	assert.EqualError(t, err, "is lower than 2.00")
	assert.Nil(t, s)
	s, err = schema.Float{Boundaries: &schema.Boundaries{Min: 2, Max: 10}}.Validate(3.1)
	assert.NoError(t, err)
	assert.Equal(t, 3.1, s)
	s, err = schema.Float{Boundaries: &schema.Boundaries{Min: math.Inf(-1), Max: 10}}.Validate(3.1)
	assert.NoError(t, err)
	assert.Equal(t, 3.1, s)
	s, err = schema.Float{Boundaries: &schema.Boundaries{Min: math.NaN(), Max: 10}}.Validate(3.1)
	assert.NoError(t, err)
	assert.Equal(t, 3.1, s)
	s, err = schema.Float{Boundaries: &schema.Boundaries{Min: 2, Max: math.Inf(1)}}.Validate(3.1)
	assert.NoError(t, err)
	assert.Equal(t, 3.1, s)
	s, err = schema.Float{Boundaries: &schema.Boundaries{Min: 2, Max: math.NaN()}}.Validate(3.1)
	assert.NoError(t, err)
	assert.Equal(t, 3.1, s)
	s, err = schema.Float{Boundaries: &schema.Boundaries{}}.Validate(1.1)
	assert.EqualError(t, err, "is greater than 0.00")
	assert.Nil(t, s)
	s, err = schema.Float{Boundaries: &schema.Boundaries{}}.Validate(-1.1)
	assert.EqualError(t, err, "is lower than 0.00")
	assert.Nil(t, s)
	s, err = schema.Float{Allowed: []float64{.1, .2, .3}}.Validate(.4)
	assert.EqualError(t, err, "not one of the allowed values")
	assert.Nil(t, s)
	s, err = schema.Float{Allowed: []float64{.1, .2, .3}}.Validate(.2)
	assert.NoError(t, err)
	assert.Equal(t, .2, s)
}

func TestFloatLesser(t *testing.T) {
	cases := []struct {
		name         string
		value, other interface{}
		expected     bool
	}{
		{`Float.Less(1.0,2.0)`, 1.0, 2.0, true},
		{`Float.Less(1.0,1.0)`, 1.0, 1.0, false},
		{`Float.Less(2.0,1.0)`, 2.0, 1.0, false},
		{`Float.Less(1.0,"2.0")`, 1.0, "2.0", false},
	}
	lessFunc := schema.Float{}.LessFunc()

	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := lessFunc(tt.value, tt.other)
			if got != tt.expected {
				t.Errorf("output for `%v`\ngot:  %v\nwant: %v", tt.name, got, tt.expected)
			}
		})
	}
}
