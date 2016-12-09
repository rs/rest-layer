package schema

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestForIncorrectlyConstructedField(t *testing.T) {
	s := Schema{
		Fields: Fields{
			"name": Field{},
		},
	}

	changes := make(map[string]interface{})
	base := make(map[string]interface{})
	base["name"] = "Fred"
	_, errs := s.Validate(changes, base)
	assert.Len(t, errs, 1, "should contain an error")
	err, ok := errs["name"][0].(error)
	assert.True(t, ok, "name field has not been set as an error")
	assert.Equal(t, err.Error(), "At least one of Validator or Schema should be set.")
}
