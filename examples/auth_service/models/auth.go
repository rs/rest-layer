package models 

import (
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/resource"
	"golang.org/x/net/context"
)

var (
	AuthSchema = schema.Schema{
		"id": schema.IDField,
		"username": schema.Field{
			Required: true,
			Validator: &schema.String{
				MaxLen: 128,
			},
		},
		"password": schema.Field{
			Required: true,
			Validator: &schema.String{
				MaxLen: 128,
			},
		},
		"access": schema.Field{
			ReadOnly: true,
			Default: false,
			Validator:  &schema.Bool{},
			OnInit: &CheckAccess,
			HookParams: []schema.HookParam{
				schema.HookParam{
					Type: schema.ConstValue,
					Param: nil,
				},
				schema.HookParam{
					Type: schema.FieldValue,
					Param: "username",
				},
				schema.HookParam{
					Type: schema.FieldValue,
					Param: "password",
				},
			},
		},
	}
)

func SetAuthUserResource(us *resource.Resource) {
	field, _ := AuthSchema["access"]
	field.HookParams[0].Param = us
}

var CheckAccess = func (value interface{}, params []interface{}) interface{} {
	users, users_ok := params[0].(*resource.Resource)
	username, u_ok := params[1].(string)
	password, p_ok := params[2].(string)

	if users_ok && u_ok && p_ok {
		l := resource.NewLookup()
		l.AddQuery(schema.Query{schema.Equal{Field: "username", Value: username}})
		list, err := users.Find(context.Background(), l, 1, 1)
		if err == nil && len(list.Items) == 1 {
			user := list.Items[0]
			if schema.VerifyPassword(user.Payload["password"], []byte(password)) {
				return true
			}
		}
	}

	return false
}
