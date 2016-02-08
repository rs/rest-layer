package models 

import "github.com/rs/rest-layer/schema"

var (
	UserSchema = schema.Schema{
		"id": schema.IDField,
		"username": schema.Field{
			Required:   true,
			Filterable: true,
			Validator: &schema.String{
				MaxLen: 128,
			},
		},
		"password": schema.PasswordField,
	}
)
