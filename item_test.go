package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewItem(t *testing.T) {
	i, err := NewItem(map[string]interface{}{"id": 1})
	assert.NoError(t, err)
	assert.Equal(t, "d2ce28b9a7fd7e4407e2b0fd499b7fe4", i.ETag)
}

func TestNewItemNoID(t *testing.T) {
	_, err := NewItem(map[string]interface{}{})
	assert.EqualError(t, err, "Missing ID field")
}

func TestNewItemNotSerializable(t *testing.T) {
	_, err := NewItem(map[string]interface{}{"id": 1, "field": func() {}})
	assert.EqualError(t, err, "json: unsupported type: func()")
}

func TestItemGetField(t *testing.T) {
	i, err := NewItem(map[string]interface{}{"id": 1, "field": map[string]interface{}{"subfield": 1}})
	assert.NoError(t, err)
	assert.Equal(t, nil, i.GetField("unknown"))
	assert.Equal(t, map[string]interface{}{"subfield": 1}, i.GetField("field"))
	assert.Equal(t, 1, i.GetField("field.subfield"))
}
