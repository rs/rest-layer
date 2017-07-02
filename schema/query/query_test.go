package query

import (
	"testing"
)

func TestMatch(t *testing.T) {
	type test struct {
		payload map[string]interface{}
		want    bool
	}
	tests := []struct {
		query string
		tests []test
	}{
		{
			`{"foo": "bar"}`, []test{
				{map[string]interface{}{"foo": "bar"}, true},
				{map[string]interface{}{"foo": "baz"}, false},
			},
		},
		{
			`{"foo": {"$ne": "bar"}}`, []test{
				{map[string]interface{}{"foo": "bar"}, false},
				{map[string]interface{}{"foo": "baz"}, true},
			},
		},
		{
			`{"foo": {"$exists": true}}`, []test{
				{map[string]interface{}{"foo": "bar"}, true},
				{map[string]interface{}{"bar": "baz"}, false},
			},
		},
		{
			`{"foo": {"$exists": false}}`, []test{
				{map[string]interface{}{"foo": "bar"}, false},
				{map[string]interface{}{"bar": "baz"}, true},
			},
		},
		{
			`{"bar": {"$gt": 1}}`, []test{
				{map[string]interface{}{"bar": 1}, false},
				{map[string]interface{}{"bar": 2}, true},
			},
		},
		{
			`{"bar": {"$gte": 2}}`, []test{
				{map[string]interface{}{"bar": 1}, false},
				{map[string]interface{}{"bar": 2}, true},
			},
		},
		{
			`{"bar": {"$lt": 2}}`, []test{
				{map[string]interface{}{"bar": 1}, true},
				{map[string]interface{}{"bar": 2}, false},
			},
		},
		{
			`{"bar": {"$lte": 1}}`, []test{
				{map[string]interface{}{"bar": 1}, true},
				{map[string]interface{}{"bar": 2}, false},
			},
		},
		{
			`{"foo": {"$in": ["bar", "baz"]}}`, []test{
				{map[string]interface{}{"foo": "bar"}, true},
				{map[string]interface{}{"foo": "foo"}, false},
			},
		},
		{
			`{"foo": {"$nin": ["bar", "baz"]}}`, []test{
				{map[string]interface{}{"foo": "bar"}, false},
				{map[string]interface{}{"foo": "foo"}, true},
			},
		},
		{
			`{"$or": [{"foo": "bar"}, {"bar": 1}]}`, []test{
				{map[string]interface{}{"foo": "bar"}, true},
				{map[string]interface{}{"foo": "foo"}, false},
				{map[string]interface{}{"bar": float64(1)}, true},
				{map[string]interface{}{"bar": "foo"}, false},
			},
		},
		{
			`{"$and": [{"foo": "bar"}, {"bar": 1}]}`, []test{
				{map[string]interface{}{"foo": "bar"}, false},
				{map[string]interface{}{"bar": float64(1)}, false},
				{map[string]interface{}{"foo": "bar", "bar": float64(1)}, true},
			},
		},
		{
			`{"foo": {"$regex": "rege[x]{1}.+some"}}`, []test{
				{map[string]interface{}{"foo": "regex-is-awesome"}, true},
			},
		},
		{
			`{"foo": {"$regex": "^(?i)my.+-rest.+$"}}`, []test{
				{map[string]interface{}{"foo": "myAwesome-RESTApplication"}, true},
			},
		},
		{
			`{"$and": [{"foo": "bar"}, {"foo": "baz"}]}`, []test{
				{map[string]interface{}{"foo": "bar"}, false},
				{map[string]interface{}{"foo": "baz"}, false},
				{map[string]interface{}{"bar": float64(1)}, false},
			},
		},
	}
	for _, tt := range tests {
		q, err := Parse(tt.query)
		if err != nil {
			t.Errorf("unexpected error for query `%v`: %v", tt.query, err)
		}
		for _, ttt := range tt.tests {
			if got := q.Match(ttt.payload); got != ttt.want {
				t.Errorf("Unexpected Match for result for query `%v` with payload %v, got %v, want %v", tt.query, ttt.payload, got, ttt.want)
			}
		}
	}
}
