package rest

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestNewRootResource(t *testing.T) {
	r := New()
	assert.Equal(t, map[string]*subResource{}, r.resources)
}

func TestRootBind(t *testing.T) {
	r := New()
	r.Bind("foo", NewResource(nil, nil, DefaultConf))
	assert.Len(t, r.resources, 1)
	log.SetOutput(ioutil.Discard)
	assert.Panics(t, func() {
		r.Bind("foo", NewResource(nil, nil, DefaultConf))
	})
}

func TestRootCompile(t *testing.T) {
	r := New()
	s := schema.Schema{"f": schema.Field{}}
	r.Bind("foo", NewResource(s, nil, DefaultConf))
	assert.NoError(t, r.Compile())
}

func TestRootCompileError(t *testing.T) {
	r := New()
	s := schema.Schema{"f": schema.Field{Validator: schema.String{Regexp: "["}}}
	r.Bind("foo", NewResource(s, nil, DefaultConf))
	assert.Error(t, r.Compile())
}

func TestRootCompileSubError(t *testing.T) {
	r := New()
	foo := r.Bind("foo", NewResource(schema.Schema{"f": schema.Field{}}, nil, DefaultConf))
	bar := foo.Bind("bar", "f", NewResource(schema.Schema{"f": schema.Field{}}, nil, DefaultConf))
	s := schema.Schema{"f": schema.Field{Validator: &schema.String{Regexp: "["}}}
	bar.Bind("baz", "f", NewResource(s, nil, DefaultConf))
	assert.EqualError(t, r.Compile(), "foo.bar.baz: schema compilation error: f: invalid regexp: error parsing regexp: missing closing ]: `[`")
}

func TestRootGetResource(t *testing.T) {
	r := New()
	r1 := NewResource(schema.Schema{"f": schema.Field{}}, nil, DefaultConf)
	r2 := NewResource(schema.Schema{"f": schema.Field{}}, nil, DefaultConf)
	foo := r.Bind("foo", r1)
	foo.Bind("bar", "f", r2)
	assert.Equal(t, r1, r.GetResource("foo"))
	assert.Equal(t, r2, r.GetResource("foo.bar"))
	assert.Nil(t, r.GetResource("foo.bar.baz"))
	assert.Nil(t, r.GetResource("bar"))
}
