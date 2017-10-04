package query

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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
			` foo ( bar : "baz" ) `,
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
			`foo(baz:true)`,
			nil,
			Projection{{Name: "foo", Params: map[string]interface{}{"baz": true}}},
		},
		{
			`foo(baz:false)`,
			nil,
			Projection{{Name: "foo", Params: map[string]interface{}{"baz": false}}},
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
			errors.New("not a string: unexpected EOF"),
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
		// Fuzz crashers
		{
			"0(0:",
			errors.New("looking for value at char 4"),
			Projection{},
		},
		{
			"0(0:0",
			errors.New("looking for `,' or ')' at char 5"),
			Projection{},
		},
		{
			"0{0(0:",
			errors.New("looking for value at char 6"),
			Projection{},
		},
		{
			"0{0(0:0",
			errors.New("looking for `,' or ')' at char 7"),
			Projection{},
		},
		{
			"o(r:0.0000000000",
			errors.New("looking for `,' or ')' at char 16"),
			Projection{},
		},
	}
	normalize := func(p string) string {
		np := make([]byte, 0, len(p))
		for _, c := range []byte(p) {
			switch c {
			case ' ', '\n', '\t':
				continue
			case '=':
				c = ':'
			}
			np = append(np, c)
		}
		return string(np)
	}
	for i := range cases {
		tc := cases[i]
		if *updateFuzzCorpus {
			os.MkdirAll("testdata/fuzz-projection/corpus", 0755)
			corpusFile := fmt.Sprintf("testdata/fuzz-projection/corpus/test%d", i)
			if err := ioutil.WriteFile(corpusFile, []byte(tc.projection), 0666); err != nil {
				t.Error(err)
			}
			continue
		}
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

			if got, want := pr.String(), normalize(tc.projection); got != want {
				t.Errorf("Projection.String:\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}
