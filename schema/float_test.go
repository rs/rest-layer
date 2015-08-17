package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloatValidator(t *testing.T) {
	s, err := Float{}.Validate(1.2)
	assert.NoError(t, err)
	assert.Equal(t, 1.2, s)
	s, err = Float{}.Validate(1)
	assert.EqualError(t, err, "not a float")
	assert.Nil(t, s)
	s, err = Float{}.Validate("1.2")
	assert.EqualError(t, err, "not a float")
	assert.Nil(t, s)
	s, err = Float{Boundaries: &Boundaries{Min: 0, Max: 2}}.Validate(3.1)
	assert.EqualError(t, err, "is greater than 2.00")
	assert.Nil(t, s)
	s, err = Float{Boundaries: &Boundaries{Min: 0, Max: 2}}.Validate(1.1)
	assert.NoError(t, err)
	assert.Equal(t, 1.1, s)
	s, err = Float{Boundaries: &Boundaries{Min: 2, Max: 10}}.Validate(1.1)
	assert.EqualError(t, err, "is lower than 2.00")
	assert.Nil(t, s)
	s, err = Float{Boundaries: &Boundaries{Min: 2, Max: 10}}.Validate(3.1)
	assert.NoError(t, err)
	assert.Equal(t, 3.1, s)
	s, err = Float{Boundaries: &Boundaries{}}.Validate(1.1)
	assert.EqualError(t, err, "is greater than 0.00")
	assert.Nil(t, s)
	s, err = Float{Boundaries: &Boundaries{}}.Validate(-1.1)
	assert.EqualError(t, err, "is lower than 0.00")
	assert.Nil(t, s)
	s, err = Float{Allowed: []float64{.1, .2, .3}}.Validate(.4)
	assert.EqualError(t, err, "not one of the allowed values")
	assert.Nil(t, s)
	s, err = Float{Allowed: []float64{.1, .2, .3}}.Validate(.2)
	assert.NoError(t, err)
	assert.Equal(t, .2, s)
}
