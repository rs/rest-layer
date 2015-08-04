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
			Sortable: true,
			Schema: &schema.Schema{
				"bar": schema.Field{Sortable: true},
			},
		},
		"baz": schema.Field{Sortable: true},
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

func TestLookupSetSortUnsortableField(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{"foo": schema.Field{Sortable: false}}
	err = l.SetSort("foo", s)
	assert.EqualError(t, err, "field is not sortable: foo")
}

func TestLookupSetSortInvalidField(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{"foo": schema.Field{Sortable: true}}
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
			Filterable: true,
			Schema: &schema.Schema{
				"bar": schema.Field{
					Validator:  schema.String{},
					Filterable: true,
				},
			},
		},
		"baz": schema.Field{
			Validator:  schema.Integer{},
			Filterable: true,
		},
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
		"foo": schema.Field{
			Validator:  schema.String{},
			Filterable: true,
		},
		"bar": schema.Field{
			Validator:  schema.Integer{},
			Filterable: true,
		},
	}
	l.SetFilter("{\"foo\": \"bar\"}", s)
	assert.True(t, l.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, l.Match(map[string]interface{}{"foo": "baz"}))
}
