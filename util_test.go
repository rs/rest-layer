package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenEtag(t *testing.T) {
	_, err := genEtag(map[string]interface{}{"id": 1, "field": func() {}})
	assert.EqualError(t, err, "json: unsupported type: func()")
	e, err := genEtag(map[string]interface{}{"id": 1, "field": "data"})
	assert.NoError(t, err)
	assert.Equal(t, "77bbb326e8b529284d96557621ca5432", e)
}

func TestGetField(t *testing.T) {
	p := map[string]interface{}{"id": 1, "field": map[string]interface{}{"subfield": 1}}
	assert.Equal(t, nil, getField(p, "unknown"))
	assert.Equal(t, map[string]interface{}{"subfield": 1}, getField(p, "field"))
	assert.Equal(t, 1, getField(p, "field.subfield"))
	assert.Equal(t, nil, getField(p, "field.subfield.subsubfield"))
}

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

func TestIsIn(t *testing.T) {
	assert.True(t, isIn("foo", "foo"))
	assert.True(t, isIn([]interface{}{"foo", "bar"}, "foo"))
	assert.False(t, isIn([]interface{}{"foo", "bar"}, "baz"))
}
