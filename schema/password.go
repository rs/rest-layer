package schema

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Password cryptes a field password using bcrypt algorithm
type Password struct {
	MaxLen int
	MinLen int
	Cost   int
	Show   bool
}

// Validate implements FieldValidator interface
func (v Password) Validate(value interface{}) (interface{}, error) {
	s, ok := value.(string)
	if !ok {
		return nil, errors.New("not a string")
	}
	if s == "$$hidden$$" {
		return nil, errors.New("passed $$hidden$$ field value back")
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

// Serialize implements FieldSerializer interface
func (v Password) Serialize(value interface{}) (interface{}, error) {
	if !v.Show {
		// Hide the field at serialization if hidden
		return "$$hidden$$", nil
	}
	return value, nil
}

// VerifyPassword compare a field of an item payload containig a hashed password
// with a clear text password and return true if they match.
func VerifyPassword(hash interface{}, password []byte) bool {
	h, ok := hash.([]byte)
	if !ok {
		return false
	}
	return bcrypt.CompareHashAndPassword(h, password) == nil
}
