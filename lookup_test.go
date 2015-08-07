package rest

import (
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestNewLookup(t *testing.T) {
	l := newLookup()
	assert.Equal(t, schema.Query{}, l.filter)
	assert.Equal(t, schema.Query{}, l.Filter())
	assert.Equal(t, []string{}, l.sort)
	assert.Equal(t, []string{}, l.Sort())
}

func TestLookupsetSort(t *testing.T) {
	var err error
	l := newLookup()
	s := schema.Schema{
		"foo": schema.Field{
			Sortable: true,
			Schema: &schema.Schema{
				"bar": schema.Field{Sortable: true},
			},
		},
		"baz": schema.Field{Sortable: true},
	}
	err = l.setSort("foo", s)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo"}, l.sort)
	err = l.setSort("foo.bar,baz", s)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo.bar", "baz"}, l.sort)
	err = l.setSort("foo.bar,-baz", s)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo.bar", "-baz"}, l.sort)
}

func TestLookupsetSortUnsortableField(t *testing.T) {
	var err error
	l := newLookup()
	s := schema.Schema{"foo": schema.Field{Sortable: false}}
	err = l.setSort("foo", s)
	assert.EqualError(t, err, "field is not sortable: foo")
}

func TestLookupsetSortInvalidField(t *testing.T) {
	var err error
	l := newLookup()
	s := schema.Schema{"foo": schema.Field{Sortable: true}}
	err = l.setSort("bar", s)
	assert.EqualError(t, err, "invalid sort field: bar")
	err = l.setSort("", s)
	assert.EqualError(t, err, "empty soft field")
	err = l.setSort("foo,", s)
	assert.EqualError(t, err, "empty soft field")
	err = l.setSort(",foo", s)
	assert.EqualError(t, err, "empty soft field")
}

func TestLookupsetFilter(t *testing.T) {
	var err error
	l := newLookup()
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
	err = l.setFilter("{\"foo\": \"bar\"}", s)
	assert.NoError(t, err)
	assert.Equal(t, schema.Query{"foo": "bar"}, l.filter)
	err = l.setFilter("{\"foo\": \"", s)
	assert.Error(t, err)
}
