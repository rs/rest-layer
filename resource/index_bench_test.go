package resource

import (
	"fmt"
	"testing"

	"github.com/rs/rest-layer/schema"
)

func BenchmarkGetResource(b *testing.B) {
	index := NewIndex()
	for i := 1; i <= 1000; i++ {
		index.Bind(fmt.Sprintf("route%04d", i), schema.Schema{}, nil, DefaultConf)
	}
	tests := map[string]string{
		"1st":      "route0001",
		"100th":    "route0100",
		"1000th":   "route1000",
		"NotFound": "notfound",
	}
	for name, rsrc := range tests {
		rsrc := rsrc // capture in context
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				index.GetResource(rsrc, nil)
			}
		})
	}
}

func BenchmarkGetResourceDepth(b *testing.B) {
	r := NewIndex()
	foo := r.Bind("foo", schema.Schema{}, nil, DefaultConf)
	bar := foo.Bind("bar", "f", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	bar.Bind("baz", "f", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	for _, p := range []string{"foo", "foo.bar", "foo.bar.baz"} {
		_, found := r.GetResource("foo", nil)
		if !found {
			b.Errorf("path %q cannot be found", p)
		}
	}
	tests := map[string]string{
		"OneLevel":    "foo",
		"TwoLevels":   "foo.bar",
		"ThreeLevels": "foo.bar.baz",
	}
	for name, rsrc := range tests {
		rsrc := rsrc // capture in context
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = r.GetResource(rsrc, nil)
			}
		})
	}
}
