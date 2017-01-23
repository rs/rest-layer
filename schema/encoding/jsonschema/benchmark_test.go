package jsonschema_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/encoding/jsonschema"
)

func BenchmarkEncoder(b *testing.B) {
	testCases := []struct {
		Name   string
		Schema schema.Schema
	}{
		// Putting the sub-benchmark with the longest name first gives better aligned output.
		{
			Name: `Schema={Fields:{"s":String{MaxLen:42}}}`,
			Schema: schema.Schema{
				Fields: schema.Fields{
					"s": {Validator: &schema.String{MaxLen: 42}},
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
			Name: `Schema={Fields:{"b":Bool{}}}`,
			Schema: schema.Schema{
				Fields: schema.Fields{
					"b": {Validator: &schema.Bool{}},
				},
			},
		},
		{
			Name:   `Schema=simple`,
			Schema: simpleSchema,
		},
		{
			Name:   `Schema=nestedObjects`,
			Schema: nestedObjectsSchema,
		},
		{
			Name:   `Schema=arrayOfObjects`,
			Schema: arrayOfObjectsSchema,
		},
		{
			Name:   `Schema=complex1`,
			Schema: complexSchema1(),
		},
		{
			Name:   `Schema=complex2`,
			Schema: complexSchema2(),
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

func complexSchema1() schema.Schema {
	mSchema := &schema.Schema{
		Description: "m",
		Fields: schema.Fields{
			"ma": String(0, 64, "ma"),
			"mb": String(0, 64, "mb"),
		},
	}
	s := schema.Schema{
		Description: "example request",
		Fields: schema.Fields{
			"a": rfc3339NanoField,
			"b": rfc3339NanoField,
			"c": String(0, 100, "b"),
			"d": Default(Integer(0, math.MaxInt64, "c"), 100),
			"e": Default(Integer(0, math.MaxInt64, "d"), 100),
			"f": Default(Integer(0, 100, "e"), 100),
			"g": Default(Integer(0, math.MaxInt64, "f"), 100),
			"h": Required(String(0, 65535, "h")),
			"i": Required(String(0, 65535, "i")),
			"j": Required(String(0, 255, "j")),
			"k": Required(String(0, 255, "k")),
			"l": Required(String(0, 65535, "l")),
			"m": schema.Field{
				Description: "m",
				Validator: &schema.Array{
					ValuesValidator: &schema.Object{
						Schema: mSchema,
					},
				},
			},
		},
	}
	Must(s.Compile())
	return s
}

func complexSchema2() schema.Schema {
	cSchema := &schema.Schema{
		Description: "c schema",
		Fields: schema.Fields{
			"ca": rfc3339NanoField,
			"cb": String(0, 64, "cb"),
			"cc": Integer(0, 65535, "cc"),
			"cd": Integer(0, 65535, "cd"),
			"ce": Integer(0, math.MaxInt64, "ce"),
			"cf": Required(String(0, 128, "cf")),
			"cg": Required(String(0, 128, "cg")),
			"ch": Integer(0, math.MaxInt64, "ch"),
			"ci": Integer(0, math.MaxInt64, "ci"),
			"cj": Integer(0, math.MaxInt64, "cj"),
			"ck": Integer(0, math.MaxInt64, "ck"),
		},
	}
	gSchema := &schema.Schema{
		Description: "gSchema",
		Fields: schema.Fields{
			"ga": Required(String(0, 64, "ga")),
			"gb": Required(String(0, 64, "gb")),
			"gc": Required(String(0, 64, "gc")),
			"gd": Required(String(0, 64, "gd")),
			"ge": Required(String(0, 64, "ge")),
		},
	}
	iSchema := &schema.Schema{
		Description: "iSchema",
		Fields: schema.Fields{
			"ia": Integer(0, math.MaxInt64, "ia"),
			"ib": Integer(0, math.MaxInt64, "ib"),
			"ic": Integer(0, math.MaxInt64, "ic"),
			"id": rfc3339NanoField,
		},
	}
	dSchema := &schema.Schema{
		Description: "dSchema",
		Fields: schema.Fields{
			"da": Required(String(0, 128, "da")),
			"db": Required(String(0, 32, "db")),
			"dc": rfc3339NanoField,
			"dd": rfc3339NanoField,
			"de": Integer(0, 99999, "de"),
			"df": Integer(0, 100, "df"),
			"dg": schema.Field{
				Description: "dg",
				Validator: &schema.Object{
					Schema: gSchema,
				},
			},
			"dh": Required(Bool("dh")),
			"di": schema.Field{
				Description: "di",
				Validator: &schema.Object{
					Schema: iSchema,
				},
			},
		},
	}
	s := schema.Schema{
		Description: "example response",
		Fields: schema.Fields{
			"a": Required(String(16, 16, "a")),
			"b": Required(String(0, 56, "b")),
			"c": schema.Field{
				Description: "c",
				Required:    true,
				Validator: &schema.Array{
					ValuesValidator: &schema.Object{
						Schema: cSchema,
					},
				},
			},
			"d": schema.Field{
				Description: "d",
				Required:    true,
				Validator: &schema.Array{
					ValuesValidator: &schema.Object{
						Schema: dSchema,
					},
				},
			},
			"e": Required(String(0, 65335, "e")),
			"f": schema.Field{
				Validator:   &schema.String{},
				Description: "f",
			},
			"g": schema.Field{
				Validator:   &schema.String{},
				Description: "g",
			},
		},
	}
	Must(s.Compile())
	return s
}
