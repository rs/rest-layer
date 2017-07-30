package query

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseProjection(t *testing.T) {
	cases := []struct {
		projection string
		err        error
		want       Projection
	}{
		{
			`foo{bar,baz}`,
			nil,
			Projection{{Name: "foo", Children: Projection{{Name: "bar"}, {Name: "baz"}}}},
		},
		{
			"  foo  \n  { \n bar \t , \n baz \t } \n",
			nil,
			Projection{{Name: "foo", Children: Projection{{Name: "bar"}, {Name: "baz"}}}},
		},
		{
			`rab:foo{bar{baz}}`,
			nil,
			Projection{{
				Name:  "foo",
				Alias: "rab",
				Children: Projection{{
					Name:     "bar",
					Children: Projection{{Name: "baz"}},
				}},
			}},
		},
		{
			`foo{rab:bar{baz}}`,
			nil,
			Projection{{
				Name: "foo",
				Children: Projection{{
					Name:     "bar",
					Alias:    "rab",
					Children: Projection{{Name: "baz"}},
				}},
			}},
		},
		{
			`foo{rab : bar{baz}}`,
			nil,
			Projection{{
				Name: "foo",
				Children: Projection{{
					Name:     "bar",
					Alias:    "rab",
					Children: Projection{{Name: "baz"}},
				}},
			}},
		},
		{
			`foo(bar:"baz")`,
			nil,
			Projection{{Name: "foo", Params: map[string]interface{}{"bar": "baz"}}},
		},
		{
			`foo(bar:"baz\"zab")`,
			nil,
			Projection{{Name: "foo", Params: map[string]interface{}{"bar": "baz\"zab"}}},
		},
		{
			`foo(bar:-0.2)`,
			nil,
			Projection{{Name: "foo", Params: map[string]interface{}{"bar": -0.2}}},
		},
		{
			`foo(bar : -0.2 , baz = "zab")`,
			nil,
			Projection{{Name: "foo", Params: map[string]interface{}{"bar": -0.2, "baz": "zab"}}},
		},
		{
			`foo{bar,baz`,
			errors.New("looking for `}' at char 11"),
			Projection{},
		},
		{
			`bar{baz}:foo`,
			errors.New("invalid char `:` at 8"),
			Projection{},
		},
		{
			`foo:{bar}`,
			errors.New("looking for field name at char 4"),
			Projection{},
		},
		{
			`foo{}`,
			errors.New("looking for field name at char 4"),
			Projection{},
		},
		{
			`{foo}`,
			errors.New("looking for field name at char 0"),
			Projection{},
		},
		{
			`,foo`,
			errors.New("looking for field name at char 0"),
			Projection{},
		},
		{
			`f oo`,
			errors.New("invalid char `o` at 2"),
			Projection{},
		},
		{
			`foo}`,
			errors.New("looking for field name and got `}' at char 3"),
			Projection{},
		},
		{
			`foo()`,
			errors.New("looking for parameter name at char 4"),
			Projection{},
		},
		{
			`foo(bar baz)`,
			errors.New("looking for : at char 8"),
			Projection{},
		},
		{
			`foo(bar`,
			errors.New("looking for : at char 7"),
			Projection{},
		},
		{
			`foo(bar:"baz)`,
			errors.New("looking for \" at char 13"),
			Projection{},
		},
		{
			`foo(bar:0a)`,
			errors.New("looking for `,' or ')' at char 9"),
			Projection{},
		},
		{
			`foo(bar:@toto)`,
			errors.New("looking for value at char 8"),
			Projection{},
		},
		{
			`foo,`,
			errors.New("looking for field name at char 4"),
			Projection{},
		},
		{
			`foo{bar,}`,
			errors.New("looking for field name at char 8"),
			Projection{},
		},
	}
	for i := range cases {
		tc := cases[i]
		t.Run(tc.projection, func(t *testing.T) {
			pr, err := ParseProjection(tc.projection)
			if !reflect.DeepEqual(err, tc.err) {
				t.Errorf("ParseProjection error:\ngot:  %v\nwant: %v", err, tc.err)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(pr, tc.want) {
				t.Errorf("Projection:\ngot:  %#v\nwant: %#v", pr, tc.want)
			}
		})
	}
}
