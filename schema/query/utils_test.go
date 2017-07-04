package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNumber(t *testing.T) {
	var ok bool
	_, ok = isNumber(1)
	assert.True(t, ok)
	_, ok = isNumber(int8(1))
	assert.True(t, ok)
	_, ok = isNumber(int16(1))
	assert.True(t, ok)
	_, ok = isNumber(int32(1))
	assert.True(t, ok)
	_, ok = isNumber(int64(1))
	assert.True(t, ok)
	_, ok = isNumber(uint(1))
	assert.True(t, ok)
	_, ok = isNumber(uint8(1))
	assert.True(t, ok)
	_, ok = isNumber(uint16(1))
	assert.True(t, ok)
	_, ok = isNumber(uint32(1))
	assert.True(t, ok)
	_, ok = isNumber(uint64(1))
	assert.True(t, ok)
	_, ok = isNumber(float32(1))
	assert.True(t, ok)
	_, ok = isNumber(float64(1))
	assert.True(t, ok)
	_, ok = isNumber("1")
	assert.False(t, ok)
}
