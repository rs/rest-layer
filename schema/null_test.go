package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullValidator(t *testing.T) {
	s, err := Null{}.Validate(nil)
	assert.NoError(t, err)
	assert.Nil(t, s)
	s, err = Null{}.Validate("null")
	assert.EqualError(t, err, "not null")
	assert.Nil(t, s)
	s, err = Null{}.Validate(0)
	assert.EqualError(t, err, "not null")
	assert.Nil(t, s)
}
