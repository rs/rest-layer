package resource

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
