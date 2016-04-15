package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourcePathValues(t *testing.T) {
	p := ResourcePath{
		&ResourcePathComponent{
			Name:  "users",
			Field: "user",
			Value: "john",
		},
		&ResourcePathComponent{
			Name:  "posts",
			Field: "id",
			Value: "123",
		},
	}
	assert.Equal(t, map[string]interface{}{"id": "123", "user": "john"}, p.Values())
}

func TestResourcePathPrepend(t *testing.T) {
	p := ResourcePath{
		&ResourcePathComponent{
			Name:  "users",
			Field: "user",
			Value: "john",
		},
	}
	p.Prepend(nil, "foo", "bar")
	assert.Equal(t, ResourcePath{
		&ResourcePathComponent{
			Field: "foo",
			Value: "bar",
		},
		&ResourcePathComponent{
			Name:  "users",
			Field: "user",
			Value: "john",
		},
	}, p)
}
