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
		index.Bind(fmt.Sprintf("route%d", i), schema.Schema{}, nil, resource.DefaultConf)
	}
	r, _ := http.NewRequest("GET", "/route800", nil)
	route, err := FindRoute(index, r)
	if err != nil {
		b.Fatal(err)
	}
	if route.Method != "GET" || route.Resource().Name() != "route800" {
		b.Fatal(route)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		route, _ = FindRoute(index, r)
		route.Release()
	}
}
