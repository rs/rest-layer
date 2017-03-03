package schema_test

import (
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestSchemaValidator(t *testing.T) {
	rc := fakeReferenceChecker{
		"foobar": {IDs: []interface{}{1, 2, 3}, Validator: &schema.Integer{}},
	}
	assert.NoError(t, rc.Compile(), "rc compile error")

	minLenSchema := &schema.Schema{
		Fields: schema.Fields{
			"foo": schema.Field{
				Validator: &schema.Bool{},
			},
			"bar": schema.Field{
				Validator: &schema.Bool{},
			},
			"baz": schema.Field{
				Validator: &schema.Bool{},
			},
		},
		MinLen: 2,
	}
	assert.NoError(t, minLenSchema.Compile(rc), "minLenSchema compile error")

	maxLenSchema := &schema.Schema{
		Fields: schema.Fields{
			"foo": schema.Field{
				Validator: &schema.Bool{},
			},
			"bar": schema.Field{
				Validator: &schema.Bool{},
			},
			"baz": schema.Field{
				Validator: &schema.Bool{},
			},
		},
		MaxLen: 2,
	}
	assert.NoError(t, maxLenSchema.Compile(rc), "maxLenSchema compile error")

	cases := []struct {
		Name                 string
		Schema               *schema.Schema
		Base, Change, Expect map[string]interface{}
		Errors               map[string][]interface{}
	}{
		{
			Name:   `MinLen=2,Validate(map[string]interface{}{"foo":true,"bar":false})`,
			Schema: minLenSchema,
			Change: map[string]interface{}{"foo": true, "bar": false},
			Expect: map[string]interface{}{"foo": true, "bar": false},
		},
		{
			Name:   `MinLen=2,Validate(map[string]interface{}{"foo":true})`,
			Schema: minLenSchema,
			Change: map[string]interface{}{"foo": true},
			Errors: map[string][]interface{}{"": []interface{}{"has fewer properties than 2"}},
		},
		{
			Name:   `MaxLen=2,Validate(map[string]interface{}{"foo":true,"bar":false})`,
			Schema: maxLenSchema,
			Change: map[string]interface{}{"foo": true, "bar": false},
			Expect: map[string]interface{}{"foo": true, "bar": false},
		},
		{
			Name:   `MaxLen=2,Validate(map[string]interface{}{"foo":true,"bar":true,"baz":false})`,
			Schema: maxLenSchema,
			Change: map[string]interface{}{"foo": true, "bar": true, "baz": false},
			Errors: map[string][]interface{}{"": []interface{}{"has more properties than 2"}},
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			doc, errs := tc.Schema.Validate(tc.Base, tc.Change)
			if len(tc.Errors) == 0 {
				assert.Len(t, errs, 0)
				assert.Equal(t, tc.Expect, doc)
			} else {
				assert.Equal(t, tc.Errors, errs)
				assert.Nil(t, doc)
			}
		})
	}
}
