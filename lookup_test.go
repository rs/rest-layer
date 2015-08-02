package rest

import (
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestNewLookup(t *testing.T) {
	l := NewLookup()
	assert.Equal(t, map[string]interface{}{}, l.Fields)
	assert.Equal(t, map[string]interface{}{}, l.Filter)
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
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, l.Filter)
	err = l.SetFilter("{\"foo.bar\": \"baz\"}", s)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"foo.bar": "baz"}, l.Filter)
	err = l.SetFilter("{\"foo\": {\"$ne\": \"bar\"}}", s)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"foo": map[string]interface{}{"$ne": "bar"}}, l.Filter)
	err = l.SetFilter("{\"baz\": {\"$gt\": 1}}", s)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"baz": map[string]interface{}{"$gt": float64(1)}}, l.Filter)
	err = l.SetFilter("{\"$or\": [{\"foo\": \"bar\"}, {\"foo\": \"baz\"}]}", s)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"$or": []interface{}{map[string]interface{}{"foo": "bar"}, map[string]interface{}{"foo": "baz"}}}, l.Filter)
	err = l.SetFilter("{\"foo\": {\"$in\": [\"bar\", \"baz\"]}}", s)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"foo": map[string]interface{}{"$in": []interface{}{"bar", "baz"}}}, l.Filter)
}

func TestLookupSetFilterUnknownField(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{}
	err = l.SetFilter("{\"unknown\": \"bar\"}", s)
	assert.EqualError(t, err, "unknown filter field: unknown")
	err = l.SetFilter("{\"unknown\": {\"$ne\": \"bar\"}}", s)
	assert.EqualError(t, err, "unknown filter field: unknown")
	err = l.SetFilter("{\"unknown\": {\"$gt\": 1}}", s)
	assert.EqualError(t, err, "unknown filter field: unknown")
	err = l.SetFilter("{\"unknown\": {\"$in\": [1, 2, 3]}}", s)
	assert.EqualError(t, err, "unknown filter field: unknown")
}

func TestLookupSetFilterInvalidType(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{
		"foo": schema.Field{Validator: schema.String{}},
		"bar": schema.Field{Validator: schema.Integer{}},
	}
	err = l.SetFilter("{", s)
	assert.EqualError(t, err, "must be a JSON object")
	err = l.SetFilter("{\"foo\": 1}", s)
	assert.EqualError(t, err, "invalid filter expression for field `foo': not a string")
	err = l.SetFilter("{\"foo\": {\"$ne\": 1}}", s)
	assert.EqualError(t, err, "invalid filter expression for field `foo': not a string")
	err = l.SetFilter("{\"bar\": {\"$gt\": 1.1}}", s)
	assert.EqualError(t, err, "invalid filter expression for field `bar': not an integer")
	err = l.SetFilter("{\"foo\": {\"$gt\": 1}}", s)
	assert.EqualError(t, err, "foo: cannot apply $gt operation on a non numerical field")
	err = l.SetFilter("{\"bar\": {\"$gt\": \"1\"}}", s)
	assert.EqualError(t, err, "bar: value for $gt must be a number")
	err = l.SetFilter("{\"bar\": {\"$in\": [\"1\"]}}", s)
	assert.EqualError(t, err, "invalid filter expression (1) for field `bar': not an integer")
	err = l.SetFilter("{\"bar\": {\"$in\": \"1\"}}", s)
	assert.EqualError(t, err, "invalid filter expression (1) for field `bar': not an integer")
	err = l.SetFilter("{\"bar\": {\"$in\": {\"bar\": \"1\"}}}", s)
	assert.EqualError(t, err, "bar: value for $in can't be a dict")
	err = l.SetFilter("{\"$or\": \"foo\"}", s)
	assert.EqualError(t, err, "value for $or must be an array of dicts")
	err = l.SetFilter("{\"$or\": [\"foo\"]}", s)
	assert.EqualError(t, err, "$or must contain at least to elements")
	err = l.SetFilter("{\"$or\": [\"foo\", \"bar\"]}", s)
	assert.EqualError(t, err, "value for $or must be an array of dicts")
	err = l.SetFilter("{\"$or\": [{\"foo\": \"bar\"}, {\"bar\": \"baz\"}]}", s)
	assert.EqualError(t, err, "invalid filter expression for field `bar': not an integer")
}

func TestLookupSetFilterInvalidHierarchy(t *testing.T) {
	var err error
	l := NewLookup()
	s := schema.Schema{
		"foo": schema.Field{Validator: schema.String{}},
		"bar": schema.Field{Validator: schema.Integer{}},
	}
	err = l.SetFilter("{\"foo\": {\"bar\": 1}}", s)
	assert.EqualError(t, err, "foo: invalid expression")
	err = l.SetFilter("{\"$ne\": \"bar\"}", s)
	assert.EqualError(t, err, "$ne can't be at first level")
	err = l.SetFilter("{\"$gt\": 1}", s)
	assert.EqualError(t, err, "$gt can't be at first level")
	err = l.SetFilter("{\"$in\": [1,2]}", s)
	assert.EqualError(t, err, "$in can't be at first level")
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
	l.SetFilter("{\"foo\": {\"$ne\": \"bar\"}}", s)
	assert.False(t, l.Match(map[string]interface{}{"foo": "bar"}))
	assert.True(t, l.Match(map[string]interface{}{"foo": "baz"}))
	l.SetFilter("{\"bar\": {\"$gt\": 1}}", s)
	assert.False(t, l.Match(map[string]interface{}{"bar": 1}))
	assert.True(t, l.Match(map[string]interface{}{"bar": 2}))
	l.SetFilter("{\"bar\": {\"$gte\": 2}}", s)
	assert.False(t, l.Match(map[string]interface{}{"bar": 1}))
	assert.True(t, l.Match(map[string]interface{}{"bar": 2}))
	l.SetFilter("{\"bar\": {\"$lt\": 2}}", s)
	assert.True(t, l.Match(map[string]interface{}{"bar": 1}))
	assert.False(t, l.Match(map[string]interface{}{"bar": 2}))
	l.SetFilter("{\"bar\": {\"$lte\": 1}}", s)
	assert.True(t, l.Match(map[string]interface{}{"bar": 1}))
	assert.False(t, l.Match(map[string]interface{}{"bar": 2}))
	l.SetFilter("{\"foo\": {\"$in\": [\"bar\", \"baz\"]}}", s)
	assert.True(t, l.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, l.Match(map[string]interface{}{"foo": "foo"}))
	l.SetFilter("{\"foo\": {\"$nin\": [\"bar\", \"baz\"]}}", s)
	assert.False(t, l.Match(map[string]interface{}{"foo": "bar"}))
	assert.True(t, l.Match(map[string]interface{}{"foo": "foo"}))
	l.SetFilter("{\"$or\": [{\"foo\": \"bar\"}, {\"bar\": 1}]}", s)
	assert.True(t, l.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, l.Match(map[string]interface{}{"foo": "foo"}))
	assert.True(t, l.Match(map[string]interface{}{"bar": float64(1)}))
	assert.False(t, l.Match(map[string]interface{}{"bar": "foo"}))
}

func TestLookupMatchFields(t *testing.T) {
	l := NewLookup()
	l.Fields["id"] = "123"
	assert.True(t, l.Match(map[string]interface{}{"id": "123"}))
	assert.False(t, l.Match(map[string]interface{}{"id": 123}))
}

func TestLookupApplyFields(t *testing.T) {
	l := NewLookup()
	l.Fields["id"] = "123"
	l.Fields["user"] = "john"
	p := map[string]interface{}{"id": "321", "name": "John Doe"}
	l.applyFields(p)
	assert.Equal(t, map[string]interface{}{"id": "123", "user": "john", "name": "John Doe"}, p)
}
