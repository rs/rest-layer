package resource

import (
	"context"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
	"github.com/stretchr/testify/assert"
)

func TestNewLookup(t *testing.T) {
	l := NewLookup()
	assert.Equal(t, query.Query{}, l.filter)
	assert.Equal(t, query.Query{}, l.Filter())
	assert.Equal(t, []string{}, l.sort)
	assert.Equal(t, []string{}, l.Sort())
}

func TestNewLookupQuery(t *testing.T) {
	l := NewLookupWithQuery(query.Query{query.Equal{Field: "foo", Value: "bar"}})
	assert.Equal(t, query.Query{query.Equal{Field: "foo", Value: "bar"}}, l.filter)
}

func TestLookupSetSort(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{
		Fields: schema.Fields{
			"foo": {
				Sortable: true,
				Schema: &schema.Schema{
					Fields: schema.Fields{
						"bar": {Sortable: true},
					},
				},
			},
			"baz": {Sortable: true},
		},
	}
	err = l.SetSort("foo", s)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo"}, l.sort)
	err = l.SetSort("foo.bar,baz", s)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo.bar", "baz"}, l.sort)
	err = l.SetSort("foo.bar,-baz", s)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo.bar", "-baz"}, l.sort)
}

func TestLookupSetSorts(t *testing.T) {
	l := NewLookup()
	l.SetSorts([]string{"foo", "bar"})
}

func TestLookupSetSortUnsortableField(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{Fields: schema.Fields{"foo": {Sortable: false}}}
	err = l.SetSort("foo", s)
	assert.EqualError(t, err, "field is not sortable: foo")
}

func TestLookupSetSortInvalidField(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{Fields: schema.Fields{"foo": {Sortable: true}}}
	err = l.SetSort("bar", s)
	assert.EqualError(t, err, "invalid sort field: bar")
	err = l.SetSort("", s)
	assert.EqualError(t, err, "empty soft field")
	err = l.SetSort("foo,", s)
	assert.EqualError(t, err, "empty soft field")
	err = l.SetSort(",foo", s)
	assert.EqualError(t, err, "empty soft field")
}

func TestLookupAddFilter(t *testing.T) {
	var err error
	l := NewLookup()
	v := schema.Schema{
		Fields: schema.Fields{
			"foo": {
				Filterable: true,
				Schema: &schema.Schema{
					Fields: schema.Fields{
						"bar": {
							Validator:  schema.String{},
							Filterable: true,
						},
					},
				},
			},
			"baz": {
				Validator:  schema.Integer{},
				Filterable: true,
			},
		},
	}
	err = l.AddFilter("{\"foo\": \"", v)
	assert.Error(t, err)
	err = l.AddFilter("{\"foo\": \"bar\"}", v)
	assert.NoError(t, err)
	assert.Equal(t, query.Query{query.Equal{Field: "foo", Value: "bar"}}, l.filter)
	err = l.AddFilter("{\"baz\": 1}", v)
	assert.NoError(t, err)
	assert.Equal(t, query.Query{query.Equal{Field: "foo", Value: "bar"}, query.Equal{Field: "baz", Value: float64(1)}}, l.filter)
	err = l.AddFilter("{\"baz\": 2}", v)
	assert.NoError(t, err)
	assert.Equal(t, query.Query{query.Equal{Field: "foo", Value: "bar"}, query.Equal{Field: "baz", Value: float64(1)}, query.Equal{Field: "baz", Value: float64(2)}}, l.filter)
}

func TestLookupAddQuery(t *testing.T) {
	l := Lookup{}
	l.AddQuery(query.Query{query.Equal{Field: "foo", Value: "bar"}})
	assert.Equal(t, query.Query{
		query.Equal{Field: "foo", Value: "bar"},
	}, l.filter)
	l.AddQuery(query.Query{query.Equal{Field: "bar", Value: "baz"}})
	assert.Equal(t, query.Query{
		query.Equal{Field: "foo", Value: "bar"},
		query.Equal{Field: "bar", Value: "baz"},
	}, l.filter)
}

func TestLookupSetSelector(t *testing.T) {
	l := NewLookup()
	v := schema.Schema{
		Fields: schema.Fields{
			"foo": {
				Schema: &schema.Schema{
					Fields: schema.Fields{
						"bar": {},
					},
				},
			},
			"baz": {},
		},
	}
	err := l.SetSelector(`foo{bar},baz`, v)
	assert.NoError(t, err)
	err = l.SetSelector(`foo,`, v)
	assert.EqualError(t, err, "looking for field name at char 4")
	err = l.SetSelector(`bar`, v)
	assert.EqualError(t, err, "bar: unknown field")
}

func TestLookupApplySelector(t *testing.T) {
	l := NewLookup()
	v := schema.Schema{
		Fields: schema.Fields{
			"foo": {
				Schema: &schema.Schema{
					Fields: schema.Fields{
						"bar": {},
					},
				},
			},
			"baz": {},
		},
	}
	ctx := context.Background()
	l.SetSelector(`foo{bar},baz`, v)
	p, err := l.ApplySelector(ctx, v, map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": "baz",
		},
	}, nil)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": "baz",
		},
	}, p)
}
