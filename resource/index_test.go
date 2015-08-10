package resource

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestNewIndex(t *testing.T) {
	r, ok := NewIndex().(*index)
	if assert.True(t, ok) {
		assert.Equal(t, map[string]*subResource{}, r.resources)
	}
}

func TestIndexBind(t *testing.T) {
	r, ok := NewIndex().(*index)
	if assert.True(t, ok) {
		r.Bind("foo", New(nil, nil, DefaultConf))
		assert.Len(t, r.resources, 1)
		log.SetOutput(ioutil.Discard)
		assert.Panics(t, func() {
			r.Bind("foo", New(nil, nil, DefaultConf))
		})
	}
}

func TestIndexCompile(t *testing.T) {
	r, ok := NewIndex().(*index)
	if assert.True(t, ok) {
		s := schema.Schema{"f": schema.Field{}}
		r.Bind("foo", New(s, nil, DefaultConf))
		assert.NoError(t, r.Compile())
	}
}

func TestIndexCompileError(t *testing.T) {
	r, ok := NewIndex().(*index)
	if assert.True(t, ok) {
		s := schema.Schema{"f": schema.Field{Validator: schema.String{Regexp: "["}}}
		r.Bind("foo", New(s, nil, DefaultConf))
		assert.Error(t, r.Compile())
	}
}

func TestIndexCompileSubError(t *testing.T) {
	r, ok := NewIndex().(*index)
	if assert.True(t, ok) {
		foo := r.Bind("foo", New(schema.Schema{"f": schema.Field{}}, nil, DefaultConf))
		bar := foo.Bind("bar", "f", New(schema.Schema{"f": schema.Field{}}, nil, DefaultConf))
		s := schema.Schema{"f": schema.Field{Validator: &schema.String{Regexp: "["}}}
		bar.Bind("baz", "f", New(s, nil, DefaultConf))
		assert.EqualError(t, r.Compile(), "foo.bar.baz: schema compilation error: f: invalid regexp: error parsing regexp: missing closing ]: `[`")
	}
}

func TestIndexGetResource(t *testing.T) {
	r := NewIndex()
	r1 := New(schema.Schema{"f": schema.Field{}}, nil, DefaultConf)
	r2 := New(schema.Schema{"f": schema.Field{}}, nil, DefaultConf)
	foo := r.Bind("foo", r1)
	foo.Bind("bar", "f", r2)
	res, field, found := r.GetResource("foo")
	assert.True(t, found)
	assert.Equal(t, r1, res)
	assert.Equal(t, "", field)
	res, field, found = r.GetResource("foo.bar")
	assert.True(t, found)
	assert.Equal(t, r2, res)
	assert.Equal(t, "f", field)
	res, field, found = r.GetResource("foo.bar.baz")
	assert.False(t, found)
	assert.Nil(t, res)
	assert.Equal(t, "", field)
	res, field, found = r.GetResource("bar")
	assert.False(t, found)
	assert.Nil(t, res)
	assert.Equal(t, "", field)
}
