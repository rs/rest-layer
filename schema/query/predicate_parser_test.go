package query

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		query string
		want  Predicate
		err   error
	}{
		{
			`{}`,
			Predicate{},
			nil,
		},
		{
			`{"foo": "bar"}`,
			Predicate{&Equal{Field: "foo", Value: "bar"}},
			nil,
		},
		{
			`{"foo": -1.1}`,
			Predicate{&Equal{Field: "foo", Value: -1.1}},
			nil,
		},
		{
			`{"foo": true}`,
			Predicate{&Equal{Field: "foo", Value: true}},
			nil,
		},
		{
			`{"foo": false}`,
			Predicate{&Equal{Field: "foo", Value: false}},
			nil,
		},
		{
			`{"foo": null}`,
			Predicate{&Equal{Field: "foo", Value: nil}},
			nil,
		},
		{
			`{"foo": "bar \n\" ❤️"}`,
			Predicate{&Equal{Field: "foo", Value: "bar \n\" ❤️"}},
			nil,
		},
		{
			`{"foo": {"bar": "baz", "baz": 1.1}}`,
			Predicate{&Equal{Field: "foo", Value: map[string]Value{"bar": "baz", "baz": 1.1}}},
			nil,
		},
		{
			`{foo: {}}`,
			Predicate{&Equal{Field: "foo", Value: map[string]Value{}}},
			nil,
		},
		{
			`{"foo.bar": "baz"}`,
			Predicate{&Equal{Field: "foo.bar", Value: "baz"}},
			nil,
		},
		{
			`{"foo": {"$ne": "bar"}}`,
			Predicate{&NotEqual{Field: "foo", Value: "bar"}},
			nil,
		},
		{
			`{"foo": {"$exists": true}}`,
			Predicate{&Exist{Field: "foo"}},
			nil,
		},
		{
			`{"foo": {"$exists": false}}`,
			Predicate{&NotExist{Field: "foo"}},
			nil,
		},
		{
			`{"baz": {"$gt": 1}}`,
			Predicate{&GreaterThan{Field: "baz", Value: float64(1)}},
			nil,
		},
		{
			`{"$or": [{"foo": "bar"}, {"foo": "baz"}]}`,
			Predicate{&Or{&Equal{Field: "foo", Value: "bar"}, &Equal{Field: "foo", Value: "baz"}}},
			nil,
		},
		{
			`{"$or": [{"foo": "bar"}, {"foo": "baz", "bar": "baz"}]}`,
			Predicate{&Or{&Equal{Field: "foo", Value: "bar"}, &And{&Equal{Field: "foo", Value: "baz"}, &Equal{Field: "bar", Value: "baz"}}}},
			nil,
		},

		{
			`{"foo": {"$regex": "regex.+awesome"}}`,
			Predicate{&Regex{Field: "foo", Value: regexp.MustCompile("regex.+awesome")}},
			nil,
		},
		{
			`{"$and": [{"foo": "bar"}, {"foo": "baz"}]}`,
			Predicate{&And{&Equal{Field: "foo", Value: "bar"}, &Equal{Field: "foo", Value: "baz"}}},
			nil,
		},
		{
			`{"$and": [{"foo": "bar", "bar": "baz"}, {"baz": "foo"}]}`,
			Predicate{&And{&And{&Equal{Field: "foo", Value: "bar"}, &Equal{Field: "bar", Value: "baz"}}, &Equal{Field: "baz", Value: "foo"}}},
			nil,
		},
		{
			`{"foo": {"$in": ["bar", "baz"]}}`,
			Predicate{&In{Field: "foo", Values: []Value{"bar", "baz"}}},
			nil,
		},
		{
			`{"foo": {"$nin": ["bar", "baz"]}}`,
			Predicate{&NotIn{Field: "foo", Values: []Value{"bar", "baz"}}},
			nil,
		},
		{
			`{ "foo" : { "$in" : [ "bar" , "baz" ] } , "bar" : { "a": [ "b", "c" ] } }`,
			Predicate{&In{Field: "foo", Values: []Value{"bar", "baz"}}, &Equal{Field: "bar", Value: map[string]Value{"a": []Value{"b", "c"}}}},
			nil,
		},
		{
			`{"bar": {"$gt": "1"}}`,
			Predicate{&GreaterThan{Field: "bar", Value: "1"}},
			nil,
		},
		{
			`{"foo": {"$elemMatch": {"bar": "one", "baz": "two"}}}`,
			Predicate{&ElemMatch{Field: "foo", Exps: []Expression{&Equal{Field: "bar", Value: "one"}, &Equal{Field: "baz", Value: "two"}}}},
			nil,
		},
		{
			`{`,
			Predicate{},
			errors.New("char 1: expected a label got '\\x00'"),
		},
		{
			`{foo: bar}`,
			Predicate{},
			errors.New("char 6: foo: unexpected char 'b'"),
		},
		{
			`{foo: "bar"`,
			Predicate{},
			errors.New("char 11: expected '}' got '\\x00'"),
		},
		{
			`{foo, "bar"`,
			Predicate{},
			errors.New("char 4: expected ':' got ','"),
		},

		{
			`{foo: "bar",}`,
			Predicate{},
			errors.New("char 12: expected a label got '}'"),
		},
		{
			`{foo: "bar"}garbage`,
			Predicate{},
			errors.New("char 12: expected EOF got 'g'"),
		},
		{
			`[]`,
			Predicate{},
			errors.New("char 0: expected '{' got '['"),
		},
		{
			`{"foo": {"$exists": 1}}`,
			Predicate{},
			errors.New("char 20: foo: $exists: not a boolean"),
		},
		{
			`{"foo": {"$exists": true`,
			Predicate{},
			errors.New("char 24: foo: $exists: expected '}' got '\\x00'"),
		},
		{
			`{"foo": {"$in": []`,
			Predicate{},
			errors.New("char 18: foo: $in: expected '}' got '\\x00'"),
		},
		{
			`{"foo": {"$ne": "bar"`,
			Predicate{},
			errors.New("char 21: foo: $ne: expected '}' got '\\x00'"),
		},
		{
			`{"foo": {"$regex": "."`,
			Predicate{},
			errors.New("char 22: foo: $regex: expected '}' got '\\x00'"),
		},
		{
			`{"foo": {"$gt": 1`,
			Predicate{},
			errors.New("char 17: foo: $gt: expected '}' got '\\x00'"),
		},
		{
			`{"foo": {"$exists`,
			Predicate{},
			errors.New("char 9: foo: invalid label: not a string: unexpected EOF"),
		},
		{
			`{"foo": {"$ne": "`,
			Predicate{},
			errors.New("char 16: foo: $ne: not a string: unexpected EOF"),
		},
		{
			`{"foo": {"$regex": "`,
			Predicate{},
			errors.New("char 19: foo: $regex: not a string: unexpected EOF"),
		},
		{
			`{"foo": "`,
			Predicate{},
			errors.New("char 8: foo: not a string: unexpected EOF"),
		},
		{
			`{"foo": nul`,
			Predicate{},
			errors.New("char 8: foo: not null"),
		},
		{
			`{"foo": {"$ne": {"bar", "baz"}}}`,
			Predicate{},
			errors.New("char 22: foo: $ne: expected ':' got ','"),
		},
		{
			`{"foo": {"$ne": {"bar": baz}}}`,
			Predicate{},
			errors.New("char 24: foo: $ne: unexpected char 'b'"),
		},
		{
			`{"foo": {"$ne": {"bar": "baz"]}}`,
			Predicate{},
			errors.New("char 29: foo: $ne: expected '}' got ']'"),
		},
		{
			`{"bar": {"$gt": 1ee0}}`,
			Predicate{},
			errors.New("char 16: bar: $gt: not a number: strconv.ParseFloat: parsing \"1ee0\": invalid syntax"),
		},

		{
			`{"bar": {"$in": {"bar": "1"}}}`,
			Predicate{},
			errors.New("char 16: bar: $in: expected '[' got '{'"),
		},
		{
			`{"bar": {"$in": ["bar": "1"]}}`,
			Predicate{},
			errors.New("char 22: bar: $in: expected ',' or ']' got ':'"),
		},
		{
			`{"bar": {"$in": [bar]}}`,
			Predicate{},
			errors.New("char 17: bar: $in: item #0: unexpected char 'b'"),
		},
		{
			`{"bar": {"$in": "bar"}}`,
			Predicate{},
			errors.New("char 16: bar: $in: expected '[' got '\"'"),
		},
		{
			`{"$or": "foo"}`,
			Predicate{},
			errors.New("char 8: $or: expected '[' got '\"'"),
		},
		{
			`{"$or": ["foo", "bar"]}`,
			Predicate{},
			errors.New("char 9: $or: expected '{' got '\"'"),
		},
		{
			`{"$or": [{"foo": "bar"}}`,
			Predicate{},
			errors.New("char 23: $or: expected ']' got '}'"),
		},
		{
			`{"$or": []}`,
			Predicate{},
			errors.New("char 10: $or: two expressions or more required"),
		},
		{
			`{"$or": [{"foo": "bar"}]}`,
			Predicate{},
			errors.New("char 24: $or: two expressions or more required"),
		},
		{
			`{"$and": "foo"}`,
			Predicate{},
			errors.New("char 9: $and: expected '[' got '\"'"),
		},
		{
			`{"$and": [{"foo": "bar"}]}`,
			Predicate{},
			errors.New("char 25: $and: two expressions or more required"),
		},
		{
			`{"$and": ["foo", "bar"]}`,
			Predicate{},
			errors.New("char 10: $and: expected '{' got '\"'"),
		},
		{
			`{"foo": {"$regex": "b[..?r"}}`,
			Predicate{},
			errors.New("char 27: foo: $regex: invalid regex: error parsing regexp: missing closing ]: `[..?r`"),
		},
		{
			`{"foo": {"$regex": "b[a-z)r"}}`,
			Predicate{},
			errors.New("char 28: foo: $regex: invalid regex: error parsing regexp: missing closing ]: `[a-z)r`"),
		},
		{
			`{"foo": {"$regex": "b(?=a)r"}}`,
			Predicate{},
			errors.New("char 28: foo: $regex: invalid regex: error parsing regexp: invalid or unsupported Perl syntax: `(?=`"),
		},
		{
			`{"foo": {"$elemMatch": "two"}}`,
			Predicate{},
			errors.New("char 23: foo: $elemMatch: expected '{' got '\"'"),
		},
		{
			`{"foo": {"$elemMatch": [{"bar": "one", "baz": "two"}]}}`,
			Predicate{},
			errors.New("char 23: foo: $elemMatch: expected '{' got '['"),
		},
		{
			`{"foo": {"$elemMatch": null}}`,
			Predicate{},
			errors.New("char 23: foo: $elemMatch: expected '{' got 'n'"),
		},
		// Hierarchy issues
		{
			`{"$ne": "bar"}`,
			Predicate{},
			errors.New("char 1: $ne: invalid placement"),
		},
		{
			`{"$exists": true}`,
			Predicate{},
			errors.New("char 1: $exists: invalid placement"),
		},
		{
			`{"$gt": 1}`,
			Predicate{},
			errors.New("char 1: $gt: invalid placement"),
		},
		{
			`{"$in": [1,2]}`,
			Predicate{},
			errors.New("char 1: $in: invalid placement"),
		},
		{
			`{"$regex": "someregexpression"}`,
			Predicate{},
			errors.New("char 1: $regex: invalid placement"),
		},
		{
			`{"$elemMatch": "someregexpression"}`,
			Predicate{},
			errors.New("char 1: $elemMatch: invalid placement"),
		},
	}
	for i := range tests {
		tt := tests[i]
		if *updateFuzzCorpus {
			os.MkdirAll("testdata/fuzz-predicate/corpus", 0755)
			corpusFile := fmt.Sprintf("testdata/fuzz-predicate/corpus/test%d", i)
			if err := ioutil.WriteFile(corpusFile, []byte(tt.query), 0666); err != nil {
				t.Error(err)
			}
			continue
		}
		t.Run(strings.Replace(tt.query, " ", "", -1), func(t *testing.T) {
			t.Parallel()
			got, err := ParsePredicate(tt.query)
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
			got, err = ParsePredicate(str)
			if err != nil {
				t.Errorf("unexpected error for reparsed query `%v`: %v", str, err)
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("invalid reparsed output for `%v`\noriginal query: `%v`\ngot:  %#v\nwant: %#v", str, tt.query, got, tt.want)
			}
		})
	}
}
