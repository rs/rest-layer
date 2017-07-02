package query

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		query string
		want  Query
		err   error
	}{
		{
			`{"foo": "bar"}`,
			Query{Equal{Field: "foo", Value: "bar"}},
			nil,
		},
		{
			`{"foo.bar": "baz"}`,
			Query{Equal{Field: "foo.bar", Value: "baz"}},
			nil,
		},
		{
			`{"foo": {"$ne": "bar"}}`,
			Query{NotEqual{Field: "foo", Value: "bar"}},
			nil,
		},
		{
			`{"foo": {"$exists": true}}`,
			Query{Exist{Field: "foo"}},
			nil,
		},
		{
			`{"foo": {"$exists": false}}`,
			Query{NotExist{Field: "foo"}},
			nil,
		},
		{
			`{"baz": {"$gt": 1}}`,
			Query{GreaterThan{Field: "baz", Value: float64(1)}},
			nil,
		},
		{
			`{"$or": [{"foo": "bar"}, {"foo": "baz"}]}`,
			Query{Or{Equal{Field: "foo", Value: "bar"}, Equal{Field: "foo", Value: "baz"}}},
			nil,
		},
		{
			`{"foo": {"$regex": "regex.+awesome"}}`,
			Query{Regex{Field: "foo", Value: regexp.MustCompile("regex.+awesome")}},
			nil,
		},
		{
			`{"$and": [{"foo": "bar"}, {"foo": "baz"}]}`,
			Query{And{Equal{Field: "foo", Value: "bar"}, Equal{Field: "foo", Value: "baz"}}},
			nil,
		},
		{
			`{"foo": {"$in": ["bar", "baz"]}}`,
			Query{In{Field: "foo", Values: []Value{"bar", "baz"}}},
			nil,
		},
		{
			`{"foo": {"$nin": ["bar", "baz"]}}`,
			Query{NotIn{Field: "foo", Values: []Value{"bar", "baz"}}},
			nil,
		},
		{
			`{`,
			Query{},
			errors.New("must be valid JSON"),
		},
		{
			`[]`,
			Query{},
			errors.New("must be a JSON object"),
		},
		{
			`{"foo": {"$exists": 1}}`,
			Query{},
			errors.New("$exists can only get Boolean as value"),
		},
		{
			`{"bar": {"$gt": "1"}}`,
			Query{},
			errors.New("bar: value for $gt must be a number"),
		},
		{
			`{"bar": {"$in": {"bar": "1"}}}`,
			Query{},
			errors.New("bar: value for $in can't be a dict"),
		},
		{
			`{"$or": "foo"}`,
			Query{},
			errors.New("value for $or must be an array of dicts"),
		},
		{
			`{"$or": ["foo"]}`,
			Query{},
			errors.New("$or must contain at least to elements"),
		},
		{
			`{"$or": ["foo", "bar"]}`,
			Query{},
			errors.New("value for $or must be an array of dicts"),
		},
		{
			`{"$and": "foo"}`,
			Query{},
			errors.New("value for $and must be an array of dicts"),
		},
		{
			`{"$and": ["foo"]}`,
			Query{},
			errors.New("$and must contain at least to elements"),
		},
		{
			`{"$and": ["foo", "bar"]}`,
			Query{},
			errors.New("value for $and must be an array of dicts"),
		},
		{
			`{"foo": {"$regex": "b[..?r"}}`,
			Query{},
			errors.New("$regex: invalid regex: error parsing regexp: missing closing ]: `[..?r`"),
		},
		{
			`{"foo": {"$regex": "b[a-z)r"}}`,
			Query{},
			errors.New("$regex: invalid regex: error parsing regexp: missing closing ]: `[a-z)r`"),
		},
		{
			`{"foo": {"$regex": "b(?=a)r"}}`,
			Query{},
			errors.New("$regex: invalid regex: error parsing regexp: invalid or unsupported Perl syntax: `(?=`"),
		},
		// Hierarchy issues
		{
			`{"foo": {"bar": 1}}`,
			Query{},
			errors.New("foo: invalid expression"),
		},
		{
			`{"$ne": "bar"}`,
			Query{},
			errors.New("$ne can't be at first level"),
		},
		{
			`{"$exists": true}`,
			Query{},
			errors.New("$exists can't be at first level"),
		},
		{
			`{"$gt": 1}`,
			Query{},
			errors.New("$gt can't be at first level"),
		},
		{
			`{"$in": [1,2]}`,
			Query{},
			errors.New("$in can't be at first level"),
		},
		{
			`{"$regex": "someregexpression"}`,
			Query{},
			errors.New("$regex can't be at first level"),
		},
	}
	for _, tt := range tests {
		got, err := Parse(tt.query)
		if !reflect.DeepEqual(err, tt.err) {
			t.Errorf("unexpected error for `%v`, got %v, want: %v", tt.query, err, tt.err)
		}
		if err == nil && !reflect.DeepEqual(got, tt.want) {
			t.Errorf("invalid output for `%v`:\ngot: %#v\nwant:%#v", tt.query, got, tt.want)
		}
	}
}
