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

func TestNewQuery(t *testing.T) {
	s := Schema{Fields: Fields{"foo": Field{Filterable: true}}}
	q, err := NewQuery(map[string]interface{}{"foo": "bar"}, s)
	assert.NoError(t, err)
	assert.Equal(t, Query{Equal{Field: "foo", Value: "bar"}}, q)
}

func TestParseQuery(t *testing.T) {
	var q Query
	var err error
	s := Schema{
		Fields: Fields{
			"foo": Field{
				Filterable: true,
				Schema: &Schema{
					Fields: Fields{
						"bar": Field{
							Validator:  String{},
							Filterable: true,
						},
					},
				},
			},
			"baz": Field{
				Validator:  Integer{},
				Filterable: true,
			},
		},
	}
	q, err = ParseQuery("{\"foo\": \"bar\"}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{Equal{Field: "foo", Value: "bar"}}, q)
	q, err = ParseQuery("{\"foo.bar\": \"baz\"}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{Equal{Field: "foo.bar", Value: "baz"}}, q)
	q, err = ParseQuery("{\"foo\": {\"$ne\": \"bar\"}}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{NotEqual{Field: "foo", Value: "bar"}}, q)
	q, err = ParseQuery("{\"foo\": {\"$exists\": true}}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{Exist{Field: "foo"}}, q)
	q, err = ParseQuery("{\"foo\": {\"$exists\": false}}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{NotExist{Field: "foo"}}, q)
	q, err = ParseQuery("{\"baz\": {\"$gt\": 1}}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{GreaterThan{Field: "baz", Value: float64(1)}}, q)
	q, err = ParseQuery("{\"$or\": [{\"foo\": \"bar\"}, {\"foo\": \"baz\"}]}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{Or{Equal{Field: "foo", Value: "bar"}, Equal{Field: "foo", Value: "baz"}}}, q)
	q, err = ParseQuery("{\"$and\": [{\"foo\": \"bar\"}, {\"foo\": \"baz\"}]}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{And{Equal{Field: "foo", Value: "bar"}, Equal{Field: "foo", Value: "baz"}}}, q)
	q, err = ParseQuery("{\"foo\": {\"$in\": [\"bar\", \"baz\"]}}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{In{Field: "foo", Values: []Value{"bar", "baz"}}}, q)
	q, err = ParseQuery("{\"foo\": {\"$nin\": [\"bar\", \"baz\"]}}", s)
	assert.NoError(t, err)
	assert.Equal(t, Query{NotIn{Field: "foo", Values: []Value{"bar", "baz"}}}, q)
}

func TestParseQueryUnfilterableField(t *testing.T) {
	var err error
	s := Schema{
		Fields: Fields{
			"foo": Field{
				Filterable: false,
				Validator:  String{},
			},
		},
	}
	_, err = ParseQuery("{\"foo\": \"bar\"}", s)
	assert.EqualError(t, err, "field is not filterable: foo")
}

func TestParseQueryUnknownField(t *testing.T) {
	var err error
	s := Schema{}
	_, err = ParseQuery("{\"unknown\": \"bar\"}", s)
	assert.EqualError(t, err, "unknown query field: unknown")
	_, err = ParseQuery("{\"unknown\": {\"$ne\": \"bar\"}}", s)
	assert.EqualError(t, err, "unknown query field: unknown")
	_, err = ParseQuery("{\"unknown\": {\"$exists\":true}}", s)
	assert.EqualError(t, err, "unknown query field: unknown")
	_, err = ParseQuery("{\"unknown\": {\"$gt\": 1}}", s)
	assert.EqualError(t, err, "unknown query field: unknown")
	_, err = ParseQuery("{\"unknown\": {\"$in\": [1, 2, 3]}}", s)
	assert.EqualError(t, err, "unknown query field: unknown")
}

func TestQueryInvalidType(t *testing.T) {
	var err error
	s := Schema{
		Fields: Fields{
			"foo": Field{Validator: String{}, Filterable: true},
			"bar": Field{Validator: Integer{}, Filterable: true},
		},
	}
	_, err = ParseQuery("{", s)
	assert.EqualError(t, err, "must be valid JSON")
	_, err = ParseQuery("[]", s)
	assert.EqualError(t, err, "must be a JSON object")
	_, err = ParseQuery("{\"foo\": 1}", s)
	assert.EqualError(t, err, "invalid query expression for field `foo': not a string")
	_, err = ParseQuery("{\"foo\": {\"$ne\": 1}}", s)
	assert.EqualError(t, err, "invalid query expression for field `foo': not a string")
	_, err = ParseQuery("{\"foo\": {\"$exists\": 1}}", s)
	assert.EqualError(t, err, "$exists can only get Boolean as value")
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
	_, err = ParseQuery("{\"$and\": \"foo\"}", s)
	assert.EqualError(t, err, "value for $and must be an array of dicts")
	_, err = ParseQuery("{\"$and\": [\"foo\"]}", s)
	assert.EqualError(t, err, "$and must contain at least to elements")
	_, err = ParseQuery("{\"$and\": [\"foo\", \"bar\"]}", s)
	assert.EqualError(t, err, "value for $and must be an array of dicts")
	_, err = ParseQuery("{\"$and\": [{\"foo\": \"bar\"}, {\"bar\": \"baz\"}]}", s)
	assert.EqualError(t, err, "invalid query expression for field `bar': not an integer")

}

func TestParseQueryInvalidHierarchy(t *testing.T) {
	var err error
	s := Schema{
		Fields: Fields{
			"foo": Field{Validator: String{}, Filterable: true},
			"bar": Field{Validator: Integer{}, Filterable: true},
		},
	}
	_, err = ParseQuery("{\"foo\": {\"bar\": 1}}", s)
	assert.EqualError(t, err, "foo: invalid expression")
	_, err = ParseQuery("{\"$ne\": \"bar\"}", s)
	assert.EqualError(t, err, "$ne can't be at first level")
	_, err = ParseQuery("{\"$exists\": true}", s)
	assert.EqualError(t, err, "$exists can't be at first level")
	_, err = ParseQuery("{\"$gt\": 1}", s)
	assert.EqualError(t, err, "$gt can't be at first level")
	_, err = ParseQuery("{\"$in\": [1,2]}", s)
	assert.EqualError(t, err, "$in can't be at first level")
}

func TestQueryMatch(t *testing.T) {
	var q Query
	s := Schema{
		Fields: Fields{
			"foo": Field{Validator: String{}, Filterable: true},
			"bar": Field{Validator: Integer{}, Filterable: true},
		},
	}
	q, _ = ParseQuery("{\"foo\": \"bar\"}", s)
	assert.True(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, q.Match(map[string]interface{}{"foo": "baz"}))
	q, _ = ParseQuery("{\"foo\": {\"$ne\": \"bar\"}}", s)
	assert.False(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.True(t, q.Match(map[string]interface{}{"foo": "baz"}))
	q, _ = ParseQuery("{\"foo\": {\"$exists\": true}}", s)
	assert.True(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, q.Match(map[string]interface{}{"bar": "baz"}))
	q, _ = ParseQuery("{\"foo\": {\"$exists\": false}}", s)
	assert.False(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.True(t, q.Match(map[string]interface{}{"bar": "baz"}))
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
	q, _ = ParseQuery("{\"$and\": [{\"foo\": \"bar\"}, {\"bar\": 1}]}", s)
	assert.False(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, q.Match(map[string]interface{}{"bar": float64(1)}))
	assert.True(t, q.Match(map[string]interface{}{"foo": "bar", "bar": float64(1)}))
	q, _ = ParseQuery("{\"$and\": [{\"foo\": \"bar\"}, {\"foo\": \"baz\"}]}", s)
	assert.False(t, q.Match(map[string]interface{}{"foo": "bar"}))
	assert.False(t, q.Match(map[string]interface{}{"foo": "baz"}))
	assert.False(t, q.Match(map[string]interface{}{"bar": float64(1)}))
}
