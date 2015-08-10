package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfDefaults(t *testing.T) {
	c := DefaultConf
	assert.Equal(t, ReadWrite, c.AllowedModes)
	assert.Equal(t, 20, c.PaginationDefaultLimit)
}

func TestModeAllowedReaWrite(t *testing.T) {
	c := Conf{AllowedModes: ReadWrite}
	assert.True(t, c.IsModeAllowed(Create))
	assert.True(t, c.IsModeAllowed(Read))
	assert.True(t, c.IsModeAllowed(Update))
	assert.True(t, c.IsModeAllowed(Replace))
	assert.True(t, c.IsModeAllowed(Delete))
	assert.True(t, c.IsModeAllowed(Clear))
	assert.True(t, c.IsModeAllowed(List))
}

func TestModeAllowedReaOnly(t *testing.T) {
	c := Conf{AllowedModes: ReadOnly}
	assert.False(t, c.IsModeAllowed(Create))
	assert.True(t, c.IsModeAllowed(Read))
	assert.False(t, c.IsModeAllowed(Update))
	assert.False(t, c.IsModeAllowed(Replace))
	assert.False(t, c.IsModeAllowed(Delete))
	assert.False(t, c.IsModeAllowed(Clear))
	assert.True(t, c.IsModeAllowed(List))
}

func TestModeAllowedWriteOnly(t *testing.T) {
	c := Conf{AllowedModes: WriteOnly}
	assert.True(t, c.IsModeAllowed(Create))
	assert.False(t, c.IsModeAllowed(Read))
	assert.True(t, c.IsModeAllowed(Update))
	assert.True(t, c.IsModeAllowed(Replace))
	assert.True(t, c.IsModeAllowed(Delete))
	assert.True(t, c.IsModeAllowed(Clear))
	assert.False(t, c.IsModeAllowed(List))
}
