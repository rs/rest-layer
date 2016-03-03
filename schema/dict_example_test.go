package schema_test

import "github.com/rs/rest-layer/schema"

func ExampleDict() {
	_ = schema.Schema{
		"dict": schema.Field{
			Validator: &schema.Dict{
				// Limit dict keys to foo and bar keys only
				KeysValidator: &schema.String{
					Allowed: []string{"foo", "bar"},
				},
				// Allow either string or integer as dict value
				ValuesValidator: &schema.AnyOf{
					0: &schema.String{},
					1: &schema.Integer{},
				},
			},
		},
	}
}
