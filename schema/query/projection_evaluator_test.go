package query

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

type resource struct {
	validator schema.Validator
}

func (r resource) Find(ctx context.Context, query *Query) ([]map[string]interface{}, error) {
	return nil, errors.New("not implemented")
}
func (r resource) MultiGet(ctx context.Context, ids []interface{}) ([]map[string]interface{}, error) {
	return nil, errors.New("not implemented")
}
func (r resource) SubResource(ctx context.Context, path string) (Resource, error) {
	return nil, errors.New("not implemented")
}
func (r resource) Validator() schema.Validator {
	return r.validator
}

func TestProjectionEval(t *testing.T) {
	s := schema.Schema{
		Fields: schema.Fields{
			"parent": {
				Schema: &schema.Schema{
					Fields: schema.Fields{
						"child": {},
					},
				},
			},
			"simple": schema.Field{},
			"with_params": {
				Params: schema.Params{
					"foo": {Validator: schema.Integer{}},
				},
				Handler: func(ctx context.Context, value interface{}, params map[string]interface{}) (interface{}, error) {
					if val, found := params["foo"]; found {
						if val == -1 {
							return nil, errors.New("some error")
						}
						return fmt.Sprintf("param is %d", val), nil
					}
					return "no param", nil
				},
			},
		},
	}

	// Basic filtering
	ctx := context.Background()
	pr := Projection{{Name: "parent", Children: Projection{{Name: "child"}}}}
	p, err := pr.Eval(ctx, map[string]interface{}{"parent": map[string]interface{}{"child": "value"}, "simple": "value"}, resource{s})
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]interface{}{"parent": map[string]interface{}{"child": "value"}}, p)
	}
	// Alias on both parent and child
	pr = Projection{{Name: "parent", Alias: "p", Children: Projection{{Name: "child", Alias: "c"}}}}
	p, err = pr.Eval(ctx, map[string]interface{}{"parent": map[string]interface{}{"child": "value"}}, resource{s})
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]interface{}{"p": map[string]interface{}{"c": "value"}}, p)
	}
	// Param call with valid value
	pr = Projection{{Name: "with_params", Params: map[string]interface{}{"foo": 1}}}
	p, err = pr.Eval(ctx, map[string]interface{}{"with_params": "value"}, resource{s})
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]interface{}{"with_params": "param is 1"}, p)
	}
	// If no param, handler do not call handler
	pr = Projection{{Name: "with_params"}}
	p, err = pr.Eval(ctx, map[string]interface{}{"with_params": "value"}, resource{s})
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]interface{}{"with_params": "value"}, p)
	}
	// Param call with valid value rejected by the handler
	pr = Projection{{Name: "with_params", Params: map[string]interface{}{"foo": -1}}}
	p, err = pr.Eval(ctx, map[string]interface{}{"with_params": "value"}, resource{s})
	assert.EqualError(t, err, "with_params: some error")
	assert.Nil(t, p)
	// Deep field lookup on a field with no child
	pr = Projection{{Name: "simple", Children: Projection{{Name: "child"}}}}
	p, err = pr.Eval(ctx, map[string]interface{}{"simple": "value"}, resource{s})
	assert.EqualError(t, err, "simple: field as no children")
	assert.Nil(t, p)
	// Deep field lookup on a field with invalid payload (no dict)
	pr = Projection{{Name: "parent", Children: Projection{{Name: "child"}}}}
	p, err = pr.Eval(ctx, map[string]interface{}{"parent": "value"}, resource{s})
	assert.EqualError(t, err, "parent: invalid value: not a dict")
	assert.Nil(t, p)
}

// func TestLookupApplySelector(t *testing.T) {
// 	l := NewLookup()
// 	v := schema.Schema{
// 		Fields: schema.Fields{
// 			"foo": {
// 				Schema: &schema.Schema{
// 					Fields: schema.Fields{
// 						"bar": {},
// 					},
// 				},
// 			},
// 			"baz": {},
// 		},
// 	}
// 	ctx := context.Background()
// 	l.SetSelector(`foo{bar},baz`, v)
// 	p, err := l.ApplySelector(ctx, v, map[string]interface{}{
// 		"foo": map[string]interface{}{
// 			"bar": "baz",
// 		},
// 	}, nil)
// 	assert.NoError(t, err)
// 	assert.Equal(t, map[string]interface{}{
// 		"foo": map[string]interface{}{
// 			"bar": "baz",
// 		},
// 	}, p)
// }
