package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolValidator(t *testing.T) {
	s, err := Bool{}.Validate(false)
	assert.NoError(t, err)
	assert.Equal(t, false, s)
	s, err = Bool{}.Validate(true)
	assert.NoError(t, err)
	assert.Equal(t, true, s)
	s, err = Bool{}.Validate("true")
	assert.EqualError(t, err, "not a Boolean")
	assert.Nil(t, s)
	s, err = Bool{}.Validate(nil)
	assert.EqualError(t, err, "not a Boolean")
	assert.Nil(t, s)
	s, err = Bool{}.Validate(0)
	assert.EqualError(t, err, "not a Boolean")
	assert.Nil(t, s)
}
