package query

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestValidateErrors(t *testing.T) {
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
			errors.New("foo: cannot apply $gt operation on a non numerical field"),
		},
		{
			`{"foo": {"$gte": 1}}`,
			errors.New("foo: cannot apply $gte operation on a non numerical field"),
		},
		{
			`{"foo": {"$lt": 1}}`,
			errors.New("foo: cannot apply $lt operation on a non numerical field"),
		},
		{
			`{"foo": {"$lte": 1}}`,
			errors.New("foo: cannot apply $lte operation on a non numerical field"),
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
		if err = q.Validate(s); !reflect.DeepEqual(err, tt.want) {
			t.Errorf("Unexpected error for `%v`:\ngot:  %v\nwant: %v", tt.query, err, tt.want)
		}
	}
}
