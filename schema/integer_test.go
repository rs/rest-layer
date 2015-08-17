package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegerValidator(t *testing.T) {
	s, err := Integer{}.Validate(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, s)
	s, err = Integer{}.Validate(1.1)
	assert.EqualError(t, err, "not an integer")
	assert.Nil(t, s)
	s, err = Integer{}.Validate("1")
	assert.EqualError(t, err, "not an integer")
	assert.Nil(t, s)
	s, err = Integer{Boundaries: &Boundaries{Min: 0, Max: 2}}.Validate(3)
	assert.EqualError(t, err, "is greater than 2")
	assert.Nil(t, s)
	s, err = Integer{Boundaries: &Boundaries{Min: 0, Max: 2}}.Validate(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, s)
	s, err = Integer{Boundaries: &Boundaries{Min: 2, Max: 10}}.Validate(1)
	assert.EqualError(t, err, "is lower than 2")
	assert.Nil(t, s)
	s, err = Integer{Boundaries: &Boundaries{Min: 2, Max: 10}}.Validate(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, s)
	s, err = Integer{Boundaries: &Boundaries{}}.Validate(1)
	assert.EqualError(t, err, "is greater than 0")
	assert.Nil(t, s)
	s, err = Integer{Boundaries: &Boundaries{}}.Validate(-1)
	assert.EqualError(t, err, "is lower than 0")
	assert.Nil(t, s)
	s, err = Integer{Allowed: []int{1, 2, 3}}.Validate(4)
	assert.EqualError(t, err, "not one of the allowed values")
	assert.Nil(t, s)
	s, err = Integer{Allowed: []int{1, 2, 3}}.Validate(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, s)
}
