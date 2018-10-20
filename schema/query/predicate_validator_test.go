package query

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/rs/rest-layer/schema"
)

func TestPrepare(t *testing.T) {
	s := schema.Schema{
		Fields: schema.Fields{
			"foo": schema.Field{Validator: &schema.String{Allowed: []string{"a", "ab"}, MinLen: 1, MaxLen: 2}, Filterable: true},
			"bar": schema.Field{Validator: &schema.Integer{Allowed: []int{1, 2}}, Filterable: true},
			"tar": schema.Field{Validator: &schema.Time{}, Filterable: true},
			"baz": schema.Field{Validator: &schema.Array{MaxLen: 1, Values: schema.Field{Validator: &schema.Time{}}}, Filterable: true},
		},
	}
	s.Compile(nil)

	now := time.Now().Format(time.RFC3339)
	nowT, _ := time.Parse(time.RFC3339, now)

	tests := []struct {
		query string
		want  Predicate
		err   error
	}{
		{
			`{"foo": "Hello"}`,
			Predicate{&Equal{Field: "foo", Value: "Hello"}},
			nil,
		},
		{
			`{"bar": 3}`,
			Predicate{&Equal{Field: "bar", Value: 3}},
			nil,
		},
		{
			`{"tar": "` + now + `"}`,
			Predicate{&Equal{Field: "tar", Value: nowT}},
			nil,
		},
		{
			`{"baz": ["` + now + `","` + now + `"]}`,
			Predicate{&Equal{Field: "baz", Value: []interface{}{nowT, nowT}}},
			nil,
		},

		{
			`{"foo": 1}`,
			Predicate{&Equal{Field: "foo", Value: nil}},
			errors.New("foo: invalid query expression: not a string"),
		},
		{
			`{"bar": "a"}`,
			Predicate{&Equal{Field: "bar", Value: nil}},
			errors.New("bar: invalid query expression: not an integer"),
		},
		{
			`{"tar": "123"}`,
			Predicate{&Equal{Field: "tar", Value: nil}},
			errors.New("tar: invalid query expression: not a time"),
		},
		{
			`{"baz": ["` + now + `","123"]}`,
			Predicate{&Equal{Field: "baz", Value: nil}},
			errors.New("baz: invalid query expression: invalid value at #2: not a time"),
		},
	}
	for _, tt := range tests {
		q, err := ParsePredicate(tt.query)
		if err != nil {
			t.Errorf("Unexpected parse error for `%v`: %v", tt.query, err)
			continue
		}
		if err = q.Prepare(s); !reflect.DeepEqual(err, tt.err) {
			t.Errorf("Unexpected error for `%v`:\ngot:  %v\nwant: %v", tt.query, err, tt.err)
		}
		if !reflect.DeepEqual(q, tt.want) {
			t.Errorf("invalid output for `%v`:\ngot:  %v\nwant: %v", tt.query, q, tt.want)
		}
	}
}

func TestPrepareErrors(t *testing.T) {
	s := schema.Schema{
		Fields: schema.Fields{
			"foo": schema.Field{Validator: schema.String{}, Filterable: true},
			"bar": schema.Field{Validator: schema.Integer{}, Filterable: true},
			"baz": schema.Field{Validator: schema.Integer{}, Filterable: false},
		},
	}
	tests := []struct {
		query string
		want  error
	}{

		{
			`{"foo": 1}`,
			errors.New("foo: invalid query expression: not a string"),
		},
		{
			`{"bar": {"$gt": 1.1}}`,
			errors.New("bar: invalid query expression: not an integer"),
		},
		{
			`{"foo": {"$ne": 1}}`,
			errors.New("foo: invalid query expression: not a string"),
		},
		{
			`{"foo": {"$gt": 1}}`,
			errors.New("foo: invalid query expression: not a string"),
		},
		{
			`{"foo": {"$gte": 1}}`,
			errors.New("foo: invalid query expression: not a string"),
		},
		{
			`{"foo": {"$lt": 1}}`,
			errors.New("foo: invalid query expression: not a string"),
		},
		{
			`{"foo": {"$lte": 1}}`,
			errors.New("foo: invalid query expression: not a string"),
		},
		{
			`{"bar": {"$in": ["1"]}}`,
			errors.New("bar: invalid query expression `\"1\"': not an integer"),
		},
		{
			`{"$or": [{"foo": "bar"}, {"bar": "baz"}]}`,
			errors.New("bar: invalid query expression: not an integer"),
		},
		{
			`{"$and": [{"foo": "bar"}, {"bar": "baz"}]}`,
			errors.New("bar: invalid query expression: not an integer"),
		},
		// Unfilterable
		{
			`{"baz": 1}`,
			errors.New("baz: field is not filterable"),
		},
		// Unknown field
		{
			`{"unknown": "bar"}`,
			errors.New("unknown: unknown query field"),
		},
		{
			`{"unknown": {"$ne": "bar"}}`,
			errors.New("unknown: unknown query field"),
		},
		{
			`{"unknown": {"$exists":true}}`,
			errors.New("unknown: unknown query field"),
		},
		{
			`{"unknown": {"$gt": 1}}`,
			errors.New("unknown: unknown query field"),
		},
		{
			`{"unknown": {"$in": [1, 2, 3]}}`,
			errors.New("unknown: unknown query field"),
		},
		{
			`{"unknown": {"$regex": "ba.+"}}`,
			errors.New("unknown: unknown query field"),
		},
	}
	for _, tt := range tests {
		q, err := ParsePredicate(tt.query)
		if err != nil {
			t.Errorf("Unexpected parse error for `%v`: %v", tt.query, err)
			continue
		}
		if err = q.Prepare(s); !reflect.DeepEqual(err, tt.want) {
			t.Errorf("Unexpected error for `%v`:\ngot:  %v\nwant: %v", tt.query, err, tt.want)
		}
	}
}
