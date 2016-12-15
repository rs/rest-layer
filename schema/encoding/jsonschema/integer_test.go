// +build go1.7

package jsonschema_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/encoding/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestIntegerValidatorNoBoundaryPanic(t *testing.T) {
	s := schema.Schema{
		Fields: schema.Fields{
			"i": schema.Field{
				Validator: &schema.Integer{},
			},
		},
	}
	assert.NotPanics(t, func() {
		enc := jsonschema.NewEncoder(new(bytes.Buffer))
		enc.Encode(&s)
	})
}

func TestIntegerValidatorEncode(t *testing.T) {
	testCases := []encoderTestCase{
		{
			name: "Allowed=nil,Boundaries=nil",
			schema: schema.Schema{
				Fields: schema.Fields{
					"i": schema.Field{
						Validator: &schema.Integer{},
					},
				},
			},
			customValidate: fieldValidator("i", `{"type": "integer"}`),
		},
		{
			name: "Allowed=[1,2,3]",
			schema: schema.Schema{
				Fields: schema.Fields{
					"i": schema.Field{
						Validator: &schema.Integer{
							Allowed: []int{1, 2, 3},
						},
					},
				},
			},
			customValidate: fieldValidator("i", `{"type": "integer", "enum": [1, 2, 3]}`),
		},
		{
			name: "Boundaries={Min:18,Max:25}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"i": schema.Field{
						Validator: &schema.Integer{
							Boundaries: &schema.Boundaries{
								Min: 18,
								Max: 25,
							},
						},
					},
				},
			},
			customValidate: fieldValidator("i", `{"type": "integer", "minimum": 18, "maximum": 25}`),
		},
		{
			name: "Boundaries={Min:18,Max:Inf}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"i": schema.Field{
						Validator: &schema.Integer{
							Boundaries: &schema.Boundaries{
								Min: 18,
								Max: math.Inf(1),
							},
						},
					},
				},
			},
			customValidate: fieldValidator("i", `{"type": "integer", "minimum": 18}`),
		},
		{
			name: "Boundaries={Min:18,Max:NaN}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"i": schema.Field{
						Validator: &schema.Integer{
							Boundaries: &schema.Boundaries{
								Min: 18,
								Max: math.NaN(),
							},
						},
					},
				},
			},
			customValidate: fieldValidator("i", `{"type": "integer", "minimum": 18}`),
		},
		{
			name: "Boundaries={Min:-Inf,Max:25}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"i": schema.Field{
						Validator: &schema.Integer{
							Boundaries: &schema.Boundaries{
								Min: math.Inf(-1),
								Max: 25,
							},
						},
					},
				},
			},
			customValidate: fieldValidator("i", `{"type": "integer", "maximum": 25}`),
		},
		{
			name: "Boundaries={Min:NaN,Max:25}",
			schema: schema.Schema{
				Fields: schema.Fields{
					"i": schema.Field{
						Validator: &schema.Integer{
							Boundaries: &schema.Boundaries{
								Min: math.NaN(),
								Max: 25,
							},
						},
					},
				},
			},
			customValidate: fieldValidator("i", `{"type": "integer", "maximum": 25}`),
		},
	}
	for i := range testCases {
		testCases[i].Run(t)
	}
}
