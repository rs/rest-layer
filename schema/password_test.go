package schema

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
)

func TestPasswordValidate(t *testing.T) {
	v, err := Password{}.Validate("secret")
	assert.NoError(t, err)
	assert.NotNil(t, v)
	v, err = Password{}.Validate([]byte("secret"))
	assert.EqualError(t, err, "not a string")
	assert.Nil(t, v)
	h, _ := bcrypt.GenerateFromPassword([]byte("secret"), 0)
	v, err = Password{}.Validate(h)
	assert.NoError(t, err)
	assert.Equal(t, h, v)
	v, err = Password{MinLen: 10}.Validate("secret")
	assert.EqualError(t, err, "is shorter than 10")
	assert.Nil(t, v)
	v, err = Password{MaxLen: 2}.Validate("secret")
	assert.EqualError(t, err, "is longer than 2")
	assert.Nil(t, v)
}

func TestVerifyPassword(t *testing.T) {
	h, _ := bcrypt.GenerateFromPassword([]byte("secret"), 0)
	assert.True(t, VerifyPassword(h, []byte("secret")))
	assert.False(t, VerifyPassword(h, []byte("wrong password")))
	assert.False(t, VerifyPassword("secret", []byte("secret")))
}
