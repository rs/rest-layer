package rest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
)

func BenchmarkFindRoute(b *testing.B) {
	index := resource.NewIndex()
	for i := 0; i < 1000; i++ {
		index.Bind(fmt.Sprintf("route%04d", i), schema.Schema{}, nil, resource.DefaultConf)
	}
	// Test
	{
		r, _ := http.NewRequest("GET", "/route0800", nil)
		route, err := FindRoute(index, r)
		if err != nil {
			b.Fatal(err)
		}
		if route.Method != "GET" || route.Resource().Name() != "route0800" {
			b.Fatal(route)
		}
	}
	b.ResetTimer()
	tests := map[string]string{
		"1st":      "/route0001",
		"100th":    "/route0100",
		"1000th":   "/route1000",
		"NotFound": "/notfound",
	}
	for name, path := range tests {
		path := path // capture in context
		b.Run(name, func(b *testing.B) {
			r, _ := http.NewRequest("GET", path, nil)
			for i := 0; i < b.N; i++ {
				route, _ := FindRoute(index, r)
				if route != nil {
					route.Release()
				}
			}
		})
	}
}
