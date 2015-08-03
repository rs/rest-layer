package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestParseQuery(t *testing.T) {
	var q Query
	var err error
	s := Schema{
		"foo": Field{
			Schema: &Schema{
				"bar": Field{Validator: String{}},
			},
		},
		"baz": Field{Validator: Integer{}},
	}
	q, err = ParseQuery("{\"foo\": \"bar\"}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{"foo": "bar"}, q)
	q, err = ParseQuery("{\"foo.bar\": \"baz\"}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{"foo.bar": "baz"}, q)
	q, err = ParseQuery("{\"foo\": {\"$ne\": \"bar\"}}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{"foo": Query{"$ne": "bar"}}, q)
	q, err = ParseQuery("{\"baz\": {\"$gt\": 1}}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{"baz": Query{"$gt": float64(1)}}, q)
	q, err = ParseQuery("{\"$or\": [{\"foo\": \"bar\"}, {\"foo\": \"baz\"}]}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{"$or": []Query{Query{"foo": "bar"}, Query{"foo": "baz"}}}, q)
	q, err = ParseQuery("{\"foo\": {\"$in\": [\"bar\", \"baz\"]}}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{"foo": Query{"$in": []interface{}{"bar", "baz"}}}, q)
}

func TestParseQueryUnknownField(t *testing.T) {
	var err error
	s := Schema{}
	_, err = ParseQuery("{\"unknown\": \"bar\"}", s)
	assert.EqualError(t, err, "unknown query field: unknown")
	_, err = ParseQuery("{\"unknown\": {\"$ne\": \"bar\"}}", s)
	assert.EqualError(t, err, "unknown query field: unknown")
	_, err = ParseQuery("{\"unknown\": {\"$gt\": 1}}", s)
	assert.EqualError(t, err, "unknown query field: unknown")
	_, err = ParseQuery("{\"unknown\": {\"$in\": [1, 2, 3]}}", s)
	assert.EqualError(t, err, "unknown query field: unknown")
}

func TestQueryInvalidType(t *testing.T) {
	var err error
	s := Schema{
		"foo": Field{Validator: String{}},
		"bar": Field{Validator: Integer{}},
	}
	_, err = ParseQuery("{", s)
	assert.EqualError(t, err, "must be a JSON object")
	_, err = ParseQuery("{\"foo\": 1}", s)
	assert.EqualError(t, err, "invalid query expression for field `foo': not a string")
	_, err = ParseQuery("{\"foo\": {\"$ne\": 1}}", s)
	assert.EqualError(t, err, "invalid query expression for field `foo': not a string")
	_, err = ParseQuery("{\"bar\": {\"$gt\": 1.1}}", s)
	assert.EqualError(t, err, "invalid query expression for field `bar': not an integer")
	_, err = ParseQuery("{\"foo\": {\"$gt\": 1}}", s)
	assert.EqualError(t, err, "foo: cannot apply $gt operation on a non numerical field")
	_, err = ParseQuery("{\"bar\": {\"$gt\": \"1\"}}", s)
	assert.EqualError(t, err, "bar: value for $gt must be a number")
	_, err = ParseQuery("{\"bar\": {\"$in\": [\"1\"]}}", s)
	assert.EqualError(t, err, "invalid query expression (1) for field `bar': not an integer")
	_, err = ParseQuery("{\"bar\": {\"$in\": \"1\"}}", s)
	assert.EqualError(t, err, "invalid query expression (1) for field `bar': not an integer")
	_, err = ParseQuery("{\"bar\": {\"$in\": {\"bar\": \"1\"}}}", s)
	assert.EqualError(t, err, "bar: value for $in can't be a dict")
	_, err = ParseQuery("{\"$or\": \"foo\"}", s)
	assert.EqualError(t, err, "value for $or must be an array of dicts")
	_, err = ParseQuery("{\"$or\": [\"foo\"]}", s)
	assert.EqualError(t, err, "$or must contain at least to elements")
	_, err = ParseQuery("{\"$or\": [\"foo\", \"bar\"]}", s)
	assert.EqualError(t, err, "value for $or must be an array of dicts")
	_, err = ParseQuery("{\"$or\": [{\"foo\": \"bar\"}, {\"bar\": \"baz\"}]}", s)
	assert.EqualError(t, err, "invalid query expression for field `bar': not an integer")
}

func TestParseQueryInvalidHierarchy(t *testing.T) {
	var err error
	s := Schema{
		"foo": Field{Validator: String{}},
		"bar": Field{Validator: Integer{}},
	}
	_, err = ParseQuery("{\"foo\": {\"bar\": 1}}", s)
	assert.EqualError(t, err, "foo: invalid expression")
	_, err = ParseQuery("{\"$ne\": \"bar\"}", s)
	assert.EqualError(t, err, "$ne can't be at first level")
	_, err = ParseQuery("{\"$gt\": 1}", s)
	assert.EqualError(t, err, "$gt can't be at first level")
	_, err = ParseQuery("{\"$in\": [1,2]}", s)
	assert.EqualError(t, err, "$in can't be at first level")
}

func TestQueryMatch(t *testing.T) {
	var q Query
	s := Schema{
		"foo": Field{Validator: String{}},
		"bar": Field{Validator: Integer{}},
	}
	q, _ = ParseQuery("{\"foo\": \"bar\"}", s)
	assert.True(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, q.Match(map[string]interface{}{"foo": "baz"}))
	q, _ = ParseQuery("{\"foo\": {\"$ne\": \"bar\"}}", s)
	assert.False(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.True(t, q.Match(map[string]interface{}{"foo": "baz"}))
	q, _ = ParseQuery("{\"bar\": {\"$gt\": 1}}", s)
	assert.False(t, q.Match(map[string]interface{}{"bar": 1}))
	assert.True(t, q.Match(map[string]interface{}{"bar": 2}))
	q, _ = ParseQuery("{\"bar\": {\"$gte\": 2}}", s)
	assert.False(t, q.Match(map[string]interface{}{"bar": 1}))
	assert.True(t, q.Match(map[string]interface{}{"bar": 2}))
	q, _ = ParseQuery("{\"bar\": {\"$lt\": 2}}", s)
	assert.True(t, q.Match(map[string]interface{}{"bar": 1}))
	assert.False(t, q.Match(map[string]interface{}{"bar": 2}))
	q, _ = ParseQuery("{\"bar\": {\"$lte\": 1}}", s)
	assert.True(t, q.Match(map[string]interface{}{"bar": 1}))
	assert.False(t, q.Match(map[string]interface{}{"bar": 2}))
	q, _ = ParseQuery("{\"foo\": {\"$in\": [\"bar\", \"baz\"]}}", s)
	assert.True(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, q.Match(map[string]interface{}{"foo": "foo"}))
	q, _ = ParseQuery("{\"foo\": {\"$nin\": [\"bar\", \"baz\"]}}", s)
	assert.False(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.True(t, q.Match(map[string]interface{}{"foo": "foo"}))
	q, _ = ParseQuery("{\"$or\": [{\"foo\": \"bar\"}, {\"bar\": 1}]}", s)
	assert.True(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, q.Match(map[string]interface{}{"foo": "foo"}))
	assert.True(t, q.Match(map[string]interface{}{"bar": float64(1)}))
	assert.False(t, q.Match(map[string]interface{}{"bar": "foo"}))
}
