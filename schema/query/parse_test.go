package query

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		query string
		want  Query
		err   error
	}{
		{
			`{}`,
			Query{},
			nil,
		},
		{
			`{"foo": "bar"}`,
			Query{Equal{Field: "foo", Value: "bar"}},
			nil,
		},
		{
			`{"foo": -1.1}`,
			Query{Equal{Field: "foo", Value: -1.1}},
			nil,
		},
		{
			`{"foo": true}`,
			Query{Equal{Field: "foo", Value: true}},
			nil,
		},
		{
			`{"foo": false}`,
			Query{Equal{Field: "foo", Value: false}},
			nil,
		},
		{
			`{"foo": "bar \n\" ❤️"}`,
			Query{Equal{Field: "foo", Value: "bar \n\" ❤️"}},
			nil,
		},
		{
			`{"foo": {"bar": "baz", "baz": 1.1}}`,
			Query{Equal{Field: "foo", Value: map[string]Value{"bar": "baz", "baz": 1.1}}},
			nil,
		},
		{
			`{foo: {}}`,
			Query{Equal{Field: "foo", Value: map[string]Value{}}},
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
			`{"$or": [{"foo": "bar"}, {"foo": "baz", "bar": "baz"}]}`,
			Query{Or{Equal{Field: "foo", Value: "bar"}, And{Equal{Field: "foo", Value: "baz"}, Equal{Field: "bar", Value: "baz"}}}},
			nil,
		},

		{
			`{"foo": {"$regex": "regex.+awesome"}}`,
			Query{Regex{Field: "foo", Value: regexp.MustCompile("regex.+awesome")}},
			nil,
		},
		{
			`{"$and": [{"foo": "bar"}, {"foo": "baz"}]}`,
			Query{Equal{Field: "foo", Value: "bar"}, Equal{Field: "foo", Value: "baz"}},
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
			`{ "foo" : { "$in" : [ "bar" , "baz" ] } , "bar" : { "a": [ "b", "c" ] } }`,
			Query{In{Field: "foo", Values: []Value{"bar", "baz"}}, Equal{Field: "bar", Value: map[string]Value{"a": []Value{"b", "c"}}}},
			nil,
		},
		{
			`{`,
			Query{},
			errors.New("char 1: expected a label got '\\x00'"),
		},
		{
			`{foo: bar}`,
			Query{},
			errors.New("char 6: foo: unexpected char 'b'"),
		},
		{
			`{foo: "bar"`,
			Query{},
			errors.New("char 11: expected '}' got '\\x00'"),
		},
		{
			`{foo, "bar"`,
			Query{},
			errors.New("char 4: expected ':' got ','"),
		},

		{
			`{foo: "bar",}`,
			Query{},
			errors.New("char 12: expected a label got '}'"),
		},
		{
			`{foo: "bar"}garbage`,
			Query{},
			errors.New("char 12: expected EOF got 'g'"),
		},
		{
			`[]`,
			Query{},
			errors.New("char 0: expected '{' got '['"),
		},
		{
			`{"foo": {"$exists": 1}}`,
			Query{},
			errors.New("char 20: foo: $exists: not a boolean"),
		},
		{
			`{"foo": {"$exists": true`,
			Query{},
			errors.New("char 24: foo: $exists: expected '}' got '\\x00'"),
		},
		{
			`{"foo": {"$in": []`,
			Query{},
			errors.New("char 18: foo: $in: expected '}' got '\\x00'"),
		},
		{
			`{"foo": {"$ne": "bar"`,
			Query{},
			errors.New("char 21: foo: $ne: expected '}' got '\\x00'"),
		},
		{
			`{"foo": {"$regex": "."`,
			Query{},
			errors.New("char 22: foo: $regex: expected '}' got '\\x00'"),
		},
		{
			`{"foo": {"$gt": 1`,
			Query{},
			errors.New("char 17: foo: $gt: expected '}' got '\\x00'"),
		},
		{
			`{"foo": {"$exists`,
			Query{},
			errors.New("char 9: foo: invalid label: not a string: unexpected EOF"),
		},
		{
			`{"foo": {"$ne": "`,
			Query{},
			errors.New("char 16: foo: $ne: not a string: unexpected EOF"),
		},
		{
			`{"foo": {"$regex": "`,
			Query{},
			errors.New("char 19: foo: $regex: not a string: unexpected EOF"),
		},
		{
			`{"foo": "`,
			Query{},
			errors.New("char 8: foo: not a string: unexpected EOF"),
		},
		{
			`{"foo": {"$ne": {"bar", "baz"}}}`,
			Query{},
			errors.New("char 22: foo: $ne: expected ':' got ','"),
		},
		{
			`{"foo": {"$ne": {"bar": baz}}}`,
			Query{},
			errors.New("char 24: foo: $ne: unexpected char 'b'"),
		},
		{
			`{"foo": {"$ne": {"bar": "baz"]}}`,
			Query{},
			errors.New("char 29: foo: $ne: expected '}' got ']'"),
		},
		{
			`{"bar": {"$gt": "1"}}`,
			Query{},
			errors.New("char 16: bar: $gt: not a number"),
		},
		{
			`{"bar": {"$gt": 1ee0}}`,
			Query{},
			errors.New("char 16: bar: $gt: not a number: strconv.ParseFloat: parsing \"1ee0\": invalid syntax"),
		},

		{
			`{"bar": {"$in": {"bar": "1"}}}`,
			Query{},
			errors.New("char 16: bar: $in: expected '[' got '{'"),
		},
		{
			`{"bar": {"$in": ["bar": "1"]}}`,
			Query{},
			errors.New("char 22: bar: $in: expected ',' or ']' got ':'"),
		},
		{
			`{"bar": {"$in": [bar]}}`,
			Query{},
			errors.New("char 17: bar: $in: item #0: unexpected char 'b'"),
		},
		{
			`{"bar": {"$in": "bar"}}`,
			Query{},
			errors.New("char 16: bar: $in: expected '[' got '\"'"),
		},
		{
			`{"$or": "foo"}`,
			Query{},
			errors.New("char 8: $or: expected '[' got '\"'"),
		},
		{
			`{"$or": ["foo", "bar"]}`,
			Query{},
			errors.New("char 9: $or: expected '{' got '\"'"),
		},
		{
			`{"$or": [{"foo": "bar"}}`,
			Query{},
			errors.New("char 23: $or: expected ']' got '}'"),
		},
		{
			`{"$or": []}`,
			Query{},
			errors.New("char 10: $or: two expressions or more required"),
		},
		{
			`{"$or": [{"foo": "bar"}]}`,
			Query{},
			errors.New("char 24: $or: two expressions or more required"),
		},
		{
			`{"$and": "foo"}`,
			Query{},
			errors.New("char 9: $and: expected '[' got '\"'"),
		},
		{
			`{"$and": [{"foo": "bar"}]}`,
			Query{},
			errors.New("char 25: $and: two expressions or more required"),
		},
		{
			`{"$and": ["foo", "bar"]}`,
			Query{},
			errors.New("char 10: $and: expected '{' got '\"'"),
		},
		{
			`{"foo": {"$regex": "b[..?r"}}`,
			Query{},
			errors.New("char 27: foo: $regex: invalid regex: error parsing regexp: missing closing ]: `[..?r`"),
		},
		{
			`{"foo": {"$regex": "b[a-z)r"}}`,
			Query{},
			errors.New("char 28: foo: $regex: invalid regex: error parsing regexp: missing closing ]: `[a-z)r`"),
		},
		{
			`{"foo": {"$regex": "b(?=a)r"}}`,
			Query{},
			errors.New("char 28: foo: $regex: invalid regex: error parsing regexp: invalid or unsupported Perl syntax: `(?=`"),
		},
		// Hierarchy issues
		{
			`{"$ne": "bar"}`,
			Query{},
			errors.New("char 1: $ne: invalid placement"),
		},
		{
			`{"$exists": true}`,
			Query{},
			errors.New("char 1: $exists: invalid placement"),
		},
		{
			`{"$gt": 1}`,
			Query{},
			errors.New("char 1: $gt: invalid placement"),
		},
		{
			`{"$in": [1,2]}`,
			Query{},
			errors.New("char 1: $in: invalid placement"),
		},
		{
			`{"$regex": "someregexpression"}`,
			Query{},
			errors.New("char 1: $regex: invalid placement"),
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(strings.Replace(tt.query, " ", "", -1), func(t *testing.T) {
			t.Parallel()
			got, err := Parse(tt.query)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("unexpected error for `%v`\ngot:  %v\nwant: %v", tt.query, err, tt.err)
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("invalid output for `%v`:\ngot:  %#v\nwant: %#v", tt.query, got, tt.want)
			}
			if err != nil {
				return
			}
			// Parse the result of query.String()
			str := got.String()
			got, err = Parse(str)
			if err != nil {
				t.Errorf("unexpected error for reparsed query `%v`: %v", str, err)
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("invalid reparsed output for `%v`\noriginal query: `%v`\ngot:  %#v\nwant: %#v", str, tt.query, got, tt.want)
			}
		})
	}
}
