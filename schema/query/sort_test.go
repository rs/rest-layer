package query

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestParseSort(t *testing.T) {
	tests := []struct {
		sort string
		want Sort
		err  error
	}{
		{"foo", Sort{SortField{Name: "foo"}}, nil},
		{"foo.bar,baz", Sort{SortField{Name: "foo.bar"}, SortField{Name: "baz"}}, nil},
		{"foo.bar,-baz", Sort{SortField{Name: "foo.bar"}, SortField{Name: "baz", Reversed: true}}, nil},
		{"", Sort{}, nil},
		{"foo,", Sort{}, errors.New("empty sort field")},
		{",foo", Sort{}, errors.New("empty sort field")},
		{",,", Sort{}, errors.New("empty sort field")},
		{"   ,   ,", Sort{}, errors.New("empty sort field")},
		{"-", Sort{}, errors.New("empty sort field")},
		{"- ", Sort{}, errors.New("empty sort field")},
	}
	for i := range tests {
		tt := tests[i]
		if *updateFuzzCorpus {
			os.MkdirAll("testdata/fuzz-sort/corpus", 0755)
			corpusFile := fmt.Sprintf("testdata/fuzz-sort/corpus/test%d", i)
			if err := ioutil.WriteFile(corpusFile, []byte(tt.sort), 0666); err != nil {
				t.Error(err)
			}
			continue
		}
		t.Run(tt.sort, func(t *testing.T) {
			got, err := ParseSort(tt.sort)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("unexpected error:\ngot:  %v\nwant: %v", err, tt.err)
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("invalid output:\ngot:  %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}

func TestSortValidate(t *testing.T) {
	s := schema.Schema{Fields: schema.Fields{
		"foo": {Sortable: false},
		"bar": {Sortable: true},
	}}
	tests := []struct {
		sort string
		err  error
	}{
		{"foo", errors.New("foo: field is not sortable")},
		{"bar", nil},
		{"baz", errors.New("baz: unknown sort field")},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.sort, func(t *testing.T) {
			sort, err := ParseSort(tt.sort)
			if err != nil {
				t.Errorf("unexpected parse error: %v", err)
			}
			if err := sort.Validate(s); !reflect.DeepEqual(err, tt.err) {
				t.Errorf("unexpected validate error:\ngot:  %#v\nwant: %#v", err, tt.err)
			}
		})
	}
}
