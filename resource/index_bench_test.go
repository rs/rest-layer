package resource

import (
	"fmt"
	"testing"

	"github.com/rs/rest-layer/schema"
)

func BenchmarkGetResource(b *testing.B) {
	index := NewIndex()
	for i := 0; i < 1000; i++ {
		index.Bind(fmt.Sprintf("route%d", i), schema.Schema{}, nil, DefaultConf)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		index.GetResource("route80", nil)
	}
}
