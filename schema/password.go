package schema

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Password crypts a field password using bcrypt algorithm.
type Password struct {
	// MinLen defines the minimum password length (default 0).
	MinLen int
	// MaxLen defines the maximum password length (default no limit).
	MaxLen int
	// Cost sets a custom bcrypt hashing cost.
	Cost int
}

var (
	// PasswordField is a common schema field for passwords. It encrypt the
	// password using bcrypt before storage and hide the value so the hash can't
	// be read back.
	PasswordField = Field{
		Description: "Write-only field storing a secret password.",
		Required:    true,
		Hidden:      true,
		Validator:   &Password{},
	}
)

// Validate implements FieldValidator interface.
func (v Password) Validate(value interface{}) (interface{}, error) {
	s, ok := value.(string)
	if !ok {
		if b, ok := value.([]byte); ok {
			// Maybe it's an already encoded version of the password.
			if _, err := bcrypt.Cost(b); err == nil {
				return b, nil
			}
		}
		return nil, errors.New("not a string")
	}
	l := len(s)
	if l < v.MinLen {
		return nil, fmt.Errorf("is shorter than %d", v.MinLen)
	}
	if v.MaxLen > 0 && l > v.MaxLen {
		return nil, fmt.Errorf("is longer than %d", v.MaxLen)
	}
	b, err := bcrypt.GenerateFromPassword([]byte(s), v.Cost)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// VerifyPassword compare a field of an item payload containing a hashed
// password with a clear text password and return true if they match.
func VerifyPassword(hash interface{}, password []byte) bool {
	h, ok := hash.([]byte)
	if !ok {
		return false
	}
	return bcrypt.CompareHashAndPassword(h, password) == nil
}
