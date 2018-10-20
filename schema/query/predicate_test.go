package query

import (
	"strings"
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestMatch(t *testing.T) {
	schemaFooInteger := schema.Schema{
		Fields: schema.Fields{
			"foo": {
				Filterable: true,
				Validator:  schema.Integer{},
			},
		},
	}
	type test struct {
		payload map[string]interface{}
		want    bool
	}
	tests := []struct {
		query  string
		tests  []test
		schema *schema.Schema
	}{
		{
			`{"foo": "bar"}`, []test{
				{map[string]interface{}{"foo": "bar"}, true},
				{map[string]interface{}{"foo": "baz"}, false},
			},
			nil,
		},
		{
			`{"foo": {"$ne": "bar"}}`, []test{
				{map[string]interface{}{"foo": "bar"}, false},
				{map[string]interface{}{"foo": "baz"}, true},
			},
			nil,
		},
		{
			`{"foo": {"$exists": true}}`, []test{
				{map[string]interface{}{"foo": "bar"}, true},
				{map[string]interface{}{"bar": "baz"}, false},
			},
			nil,
		},
		{
			`{"foo": {"$exists": false}}`, []test{
				{map[string]interface{}{"foo": "bar"}, false},
				{map[string]interface{}{"bar": "baz"}, true},
			},
			nil,
		},
		{
			`{"foo": {"$gt": 1}}`, []test{
				{map[string]interface{}{"foo": 1}, false},
				{map[string]interface{}{"foo": 2}, true},
			},
			&schemaFooInteger,
		},
		{
			`{"foo": {"$gte": 2}}`, []test{
				{map[string]interface{}{"foo": 1}, false},
				{map[string]interface{}{"foo": 2}, true},
			},
			&schemaFooInteger,
		},
		{
			`{"foo": {"$lt": 2}}`, []test{
				{map[string]interface{}{"foo": 1}, true},
				{map[string]interface{}{"foo": 2}, false},
			},
			&schemaFooInteger,
		},
		{
			`{"foo": {"$lte": 1}}`, []test{
				{map[string]interface{}{"foo": 1}, true},
				{map[string]interface{}{"foo": 2}, false},
			},
			&schemaFooInteger,
		},
		{
			`{"foo": {"$in": ["bar", "baz"]}}`, []test{
				{map[string]interface{}{"foo": "bar"}, true},
				{map[string]interface{}{"foo": "foo"}, false},
			},
			nil,
		},
		{
			`{"foo": {"$in": ["baz"]}}`, []test{
				{map[string]interface{}{"foo": []interface{}{"baz"}}, true},
				{map[string]interface{}{"foo": []interface{}{"bar"}}, false},
			},
			nil,
		},
		{
			`{"foo": {"$nin": ["bar", "baz"]}}`, []test{
				{map[string]interface{}{"foo": "bar"}, false},
				{map[string]interface{}{"foo": "foo"}, true},
			},
			nil,
		},
		{
			`{"foo": {"$nin": ["baz"]}}`, []test{
				{map[string]interface{}{"foo": []interface{}{"baz"}}, false},
				{map[string]interface{}{"foo": []interface{}{"bar"}}, true},
			},
			nil,
		},
		{
			`{"$or": [{"foo": "bar"}, {"bar": 1}]}`, []test{
				{map[string]interface{}{"foo": "bar"}, true},
				{map[string]interface{}{"foo": "foo"}, false},
				{map[string]interface{}{"bar": float64(1)}, true},
				{map[string]interface{}{"bar": "foo"}, false},
			},
			nil,
		},
		{
			`{"$and": [{"foo": "bar"}, {"bar": 1}]}`, []test{
				{map[string]interface{}{"foo": "bar"}, false},
				{map[string]interface{}{"bar": float64(1)}, false},
				{map[string]interface{}{"foo": "bar", "bar": float64(1)}, true},
			},
			nil,
		},
		{
			`{"foo": {"$regex": "rege[x]{1}.+some"}}`, []test{
				{map[string]interface{}{"foo": "regex-is-awesome"}, true},
			},
			nil,
		},
		{
			`{"foo": {"$regex": "^(?i)my.+-rest.+$"}}`, []test{
				{map[string]interface{}{"foo": "myAwesome-RESTApplication"}, true},
			},
			nil,
		},
		{
			`{"$and": [{"foo": "bar"}, {"foo": "baz"}]}`, []test{
				{map[string]interface{}{"foo": "bar"}, false},
				{map[string]interface{}{"foo": "baz"}, false},
				{map[string]interface{}{"bar": float64(1)}, false},
			},
			nil,
		},
		{
			`{"foo.bar": "baz"}`, []test{
				{map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}, true},
				{map[string]interface{}{"foo": map[string]interface{}{"bar": "bar"}}, false},
			},
			nil,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(strings.Replace(tt.query, " ", "", -1), func(t *testing.T) {
			t.Parallel()
			q, err := ParsePredicate(tt.query)
			if err != nil {
				t.Errorf("Unexpected error for query `%v`: %v", tt.query, err)
			}
			// use the schema to prepare expressions only when it is needed/defined
			if tt.schema != nil {
				err := q.Prepare(tt.schema)
				if err != nil {
					t.Errorf("Unexpected error for query `%v`: %v", tt.query, err)
				}
			}
			for _, ttt := range tt.tests {
				if got := q.Match(ttt.payload); got != ttt.want {
					t.Errorf("Unexpected Match for result for query `%v` with payload %v, got %v, want %v", tt.query, ttt.payload, got, ttt.want)
				}
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := map[string]string{
		`{"foo": "bar"}`:                                          `{foo: "bar"}`,
		`{"foo": {"$ne": "bar"}}`:                                 `{foo: {$ne: "bar"}}`,
		`{"foo": {"$exists": true}}`:                              `{foo: {$exists: true}}`,
		`{"foo": {"$exists": false}}`:                             `{foo: {$exists: false}}`,
		`{"bar": {"$gt": 1}}`:                                     `{bar: {$gt: 1}}`,
		`{"bar": {"$gte": 2}}`:                                    `{bar: {$gte: 2}}`,
		`{"bar": {"$lt": 2}}`:                                     `{bar: {$lt: 2}}`,
		`{"bar": {"$lte": 1}}`:                                    `{bar: {$lte: 1}}`,
		`{"foo": {"$in": ["bar", "baz"]}}`:                        `{foo: {$in: ["bar", "baz"]}}`,
		`{"foo": {"$nin": ["bar", "baz"]}}`:                       `{foo: {$nin: ["bar", "baz"]}}`,
		`{"$or": [{"foo": "bar"}, {"bar": 1}]}`:                   `{$or: [{foo: "bar"}, {bar: 1}]}`,
		`{"$and": [{"foo": "bar"}, {"bar": 1}]}`:                  `{$and: [{foo: "bar"}, {bar: 1}]}`,
		`{"foo": {"$regex": "rege[x]{1}.+some"}}`:                 `{foo: {$regex: "rege[x]{1}.+some"}}`,
		`{"foo": {"$regex": "^(?i)my.+-rest.+$"}}`:                `{foo: {$regex: "^(?i)my.+-rest.+$"}}`,
		`{"$and": [{"foo": "bar"}, {"foo": "baz"}]}`:              `{$and: [{foo: "bar"}, {foo: "baz"}]}`,
		`{"foo": "bar", "$or": [{"bar": "baz"}, {"bar": "foo"}]}`: `{foo: "bar", $or: [{bar: "baz"}, {bar: "foo"}]}`,
		`{"foo": ["bar", "baz"]}`:                                 `{foo: ["bar","baz"]}`,
		`{"foo.bar": "baz"}`:                                      `{foo.bar: "baz"}`,
	}
	for query, want := range tests {
		q, err := ParsePredicate(query)
		if err != nil {
			t.Errorf("Unexpected error for query `%v`: %v", query, err)
		}
		if got := q.String(); got != want {
			t.Errorf("Unexpected String result for `%v`\ngot:  `%v`\nwant: `%v`", query, got, want)
		}
	}
}
