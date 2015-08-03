package rest

import (
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestNewLookup(t *testing.T) {
	l := NewLookup()
	assert.Equal(t, schema.Query{}, l.Filter)
	assert.Equal(t, []string{}, l.Sort)
}

func TestLookupSetSort(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{
		"foo": schema.Field{
			Schema: &schema.Schema{
				"bar": schema.Field{},
			},
		},
		"baz": schema.Field{},
	}
	err = l.SetSort("foo", s)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo"}, l.Sort)
	err = l.SetSort("foo.bar,baz", s)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo.bar", "baz"}, l.Sort)
	err = l.SetSort("foo.bar,-baz", s)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo.bar", "-baz"}, l.Sort)
}

func TestLookupSetSortInvalidField(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{"foo": schema.Field{}}
	err = l.SetSort("bar", s)
	assert.EqualError(t, err, "invalid sort field: bar")
	err = l.SetSort("", s)
	assert.EqualError(t, err, "empty soft field")
	err = l.SetSort("foo,", s)
	assert.EqualError(t, err, "empty soft field")
	err = l.SetSort(",foo", s)
	assert.EqualError(t, err, "empty soft field")
}

func TestLookupSetFilter(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{
		"foo": schema.Field{
			Schema: &schema.Schema{
				"bar": schema.Field{Validator: schema.String{}},
			},
		},
		"baz": schema.Field{Validator: schema.Integer{}},
	}
	err = l.SetFilter("{\"foo\": \"bar\"}", s)
	assert.NoError(t, err)
	assert.Equal(t, schema.Query{"foo": "bar"}, l.Filter)
	err = l.SetFilter("{\"foo\": \"", s)
	assert.Error(t, err)
}

func TestLookupMatch(t *testing.T) {
	l := NewLookup()
	s := schema.Schema{
		"foo": schema.Field{Validator: schema.String{}},
		"bar": schema.Field{Validator: schema.Integer{}},
	}
	l.SetFilter("{\"foo\": \"bar\"}", s)
	assert.True(t, l.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, l.Match(map[string]interface{}{"foo": "baz"}))
}
