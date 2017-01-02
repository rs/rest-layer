package jsonschema_test

import (
	"bytes"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/encoding/jsonschema"
)

func BenchmarkEncoder(b *testing.B) {
	testCases := []struct {
		Name   string
		Schema schema.Schema
	}{
		{
			Name: `Schema={Fields:{"b":Bool{}}}`,
			Schema: schema.Schema{
				Fields: schema.Fields{
					"b": {Validator: &schema.Bool{}},
				},
			},
		},
		{
			Name: `Schema={Fields:{"s":String{}}}`,
			Schema: schema.Schema{
				Fields: schema.Fields{
					"s": {Validator: &schema.String{}},
				},
			},
		},
		{
			Name: `Schema={Fields:{"s":String{MaxLen:42}}}`,
			Schema: schema.Schema{
				Fields: schema.Fields{
					"s": {Validator: &schema.String{MaxLen: 42}},
				},
			},
		},
		{
			Name:   `Schema=Simple`,
			Schema: studentSchema,
		},
		{
			Name:   `Schema=Complex1`,
			Schema: getComplexSchema1(),
		},
		{
			Name:   `Schema=Complex2`,
			Schema: getComplexSchema2(),
		},
	}
	for i := range testCases {
		buf := bytes.NewBuffer(make([]byte, 2<<19))
		tc := testCases[i]
		b.Run(tc.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf.Truncate(0)
				enc := jsonschema.NewEncoder(buf)
				enc.Encode(&tc.Schema)
			}
		})
	}
}
