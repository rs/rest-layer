package jsonschema_test

import (
	"github.com/rs/rest-layer/schema"
	"math"
	"time"
)

var (
	rfc3339NanoDefault = schema.Field{
		Description: "UTC start time in RFC3339 format with Nano second support, e.g. 2006-01-02T15:04:05.999999999Z",
		Validator: &schema.Time{
			TimeLayouts: []string{time.RFC3339Nano},
		},
	}
)

// Default modifies an existing schema.Field with a default value.
func Default(f schema.Field, v interface{}) schema.Field {
	f.Default = v
	return f
}

// Required modifies an existing schema.Field to be required.
func Required(s schema.Field) schema.Field {
	s.Required = true
	return s
}

// RFC3339Nano schema.Field template
func RFC3339Nano() schema.Field {
	return rfc3339NanoDefault
}

// String field
func String(min, max int, description string) schema.Field {
	return schema.Field{
		Description: description,
		Default:     "",
		Validator: &schema.String{
			MinLen: min,
			MaxLen: max,
		},
	}
}

// Integer returns schema.Field template for int validation.
func Integer(min, max int, description string) schema.Field {
	return schema.Field{
		Default:     min,
		Description: description,
		Validator: &schema.Integer{
			Boundaries: &schema.Boundaries{
				Min: float64(min),
				Max: float64(max),
			},
		},
	}
}

// Bool validator helper
func Bool(description string) schema.Field {
	return schema.Field{
		Description: description,
		Validator:   &schema.Bool{},
	}
}

func getComplexSchema1() schema.Schema {
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
			"a": RFC3339Nano(),
			"b": RFC3339Nano(),
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
	s.Compile()
	return s
}

func getComplexSchema2() schema.Schema {
	cSchema := &schema.Schema{
		Description: "c schema",
		Fields: schema.Fields{
			"ca": RFC3339Nano(),
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
			"id": RFC3339Nano(),
		},
	}
	dSchema := &schema.Schema{
		Description: "dSchema",
		Fields: schema.Fields{
			"da": Required(String(0, 128, "da")),
			"db": Required(String(0, 32, "db")),
			"dc": RFC3339Nano(),
			"dd": RFC3339Nano(),
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
	s.Compile()
	return s
}
