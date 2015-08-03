package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteApplyFields(t *testing.T) {
	r := route{
		fields: map[string]interface{}{
			"id":   "123",
			"user": "john",
		},
	}
	p := map[string]interface{}{"id": "321", "name": "John Doe"}
	r.applyFields(p)
	assert.Equal(t, map[string]interface{}{"id": "123", "user": "john", "name": "John Doe"}, p)
}
