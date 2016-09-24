package jsonschema_test

import (
	"math"
	"testing"

	"github.com/rs/rest-layer/schema"
)

func TestFloatValidatorEncode(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: "Allowed=nil,Boundaries=nil",
			schema: schema.Schema{
				Fields: schema.Fields{
					"f": schema.Field{
						Validator: &schema.Float{},
					},
				},
			},
			customValidate: fieldValidator("f", `{"type": "number"}`),
		},
		{
			name: "Allowed=[0,0.5,100]",
			schema: schema.Schema{
				Fields: schema.Fields{
					"f": schema.Field{
						Validator: &schema.Float{
							Allowed: []float64{0, 0.5, 100},
						},
					},
				},
			},
			customValidate: fieldValidator("f", `{"type": "number", "enum": [0,0.5,100]}`),
		},
		{
			name: "Boundaries={Min:0,Max:100}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"f": schema.Field{
						Validator: &schema.Float{
							Boundaries: &schema.Boundaries{
								Min: 0,
								Max: 100,
							},
						},
					},
				},
			},
			customValidate: fieldValidator("f", `{"type": "number", "minimum": 0, "maximum": 100}`),
		},
		{
			name: "Boundaries={Min:0,Max:Inf}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"f": schema.Field{
						Validator: &schema.Float{
							Boundaries: &schema.Boundaries{
								Min: 0,
								Max: math.Inf(1),
							},
						},
					},
				},
			},
			customValidate: fieldValidator("f", `{"type": "number", "minimum": 0}`),
		},
		{
			name: "Boundaries={Min:0,Max:NaN}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"f": schema.Field{
						Validator: &schema.Float{
							Boundaries: &schema.Boundaries{
								Min: 0,
								Max: math.NaN(),
							},
						},
					},
				},
			},
			customValidate: fieldValidator("f", `{"type": "number", "minimum": 0}`),
		},
		{
			name: "Boundaries={Min:-Inf,Max:100}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"f": schema.Field{
						Validator: &schema.Float{
							Boundaries: &schema.Boundaries{
								Min: math.Inf(-1),
								Max: 100,
							},
						},
					},
				},
			},
			customValidate: fieldValidator("f", `{"type": "number", "maximum": 100}`),
		},
		{
			name: "Boundaries={Min:NaN,Max:100}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"f": schema.Field{
						Validator: &schema.Float{
							Boundaries: &schema.Boundaries{
								Min: math.NaN(),
								Max: 100,
							},
						},
					},
				},
			},
			customValidate: fieldValidator("f", `{"type": "number", "maximum": 100}`),
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
