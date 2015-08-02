package rest

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
	assert.True(t, c.isModeAllowed(Create))
	assert.True(t, c.isModeAllowed(Read))
	assert.True(t, c.isModeAllowed(Update))
	assert.True(t, c.isModeAllowed(Replace))
	assert.True(t, c.isModeAllowed(Delete))
	assert.True(t, c.isModeAllowed(Clear))
	assert.True(t, c.isModeAllowed(List))
}

func TestModeAllowedReaOnly(t *testing.T) {
	c := Conf{AllowedModes: ReadOnly}
	assert.False(t, c.isModeAllowed(Create))
	assert.True(t, c.isModeAllowed(Read))
	assert.False(t, c.isModeAllowed(Update))
	assert.False(t, c.isModeAllowed(Replace))
	assert.False(t, c.isModeAllowed(Delete))
	assert.False(t, c.isModeAllowed(Clear))
	assert.True(t, c.isModeAllowed(List))
}

func TestModeAllowedWriteOnly(t *testing.T) {
	c := Conf{AllowedModes: WriteOnly}
	assert.True(t, c.isModeAllowed(Create))
	assert.False(t, c.isModeAllowed(Read))
	assert.True(t, c.isModeAllowed(Update))
	assert.True(t, c.isModeAllowed(Replace))
	assert.True(t, c.isModeAllowed(Delete))
	assert.True(t, c.isModeAllowed(Clear))
	assert.False(t, c.isModeAllowed(List))
}
