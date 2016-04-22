package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReferenceValidate(t *testing.T) {
	v, err := Reference{}.Validate("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", v)
}
