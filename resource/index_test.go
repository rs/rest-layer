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
		assert.Equal(t, subResources{}, r.resources)
	}
}

func TestIndexBind(t *testing.T) {
	r, ok := NewIndex().(*index)
	if assert.True(t, ok) {
		r.Bind("foo", schema.Schema{}, nil, DefaultConf)
		assert.Len(t, r.resources, 1)
		log.SetOutput(ioutil.Discard)
		assert.Panics(t, func() {
			r.Bind("foo", schema.Schema{}, nil, DefaultConf)
		})
	}
}

func TestIndexCompile(t *testing.T) {
	r, ok := NewIndex().(*index)
	if !assert.True(t, ok) {
		return
	}
	s := schema.Schema{Fields: schema.Fields{"f": {}}}
	r.Bind("foo", s, nil, DefaultConf)
	assert.NoError(t, r.Compile())
}

func TestIndexCompileError(t *testing.T) {
	r, ok := NewIndex().(*index)
	if !assert.True(t, ok) {
		return
	}
	s := schema.Schema{
		Fields: schema.Fields{
			"f": {Validator: schema.String{Regexp: "["}},
		},
	}
	r.Bind("foo", s, nil, DefaultConf)
	assert.Error(t, r.Compile())
}

func TestIndexCompileSubError(t *testing.T) {
	r, ok := NewIndex().(*index)
	if !assert.True(t, ok) {
		return
	}
	foo := r.Bind("foo", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	bar := foo.Bind("bar", "f", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	s := schema.Schema{Fields: schema.Fields{"f": {Validator: &schema.String{Regexp: "["}}}}
	bar.Bind("baz", "f", s, nil, DefaultConf)
	assert.EqualError(t, r.Compile(), "foo.bar.baz: schema compilation error: f: invalid regexp: error parsing regexp: missing closing ]: `[`")
}

func TestIndexCompileReferenceChecker(t *testing.T) {
	i, ok := NewIndex().(*index)
	if !assert.True(t, ok) {
		return
	}

	i.Bind("b", schema.Schema{Fields: schema.Fields{"id": {}}}, nil, DefaultConf)
	i.Bind("a", schema.Schema{Fields: schema.Fields{"ref": {
		Validator: &schema.Reference{Path: "b"},
	}}}, nil, DefaultConf)
	assert.NoError(t, i.Compile())
}

func TestIndexCompileReferenceCheckerError(t *testing.T) {
	i, ok := NewIndex().(*index)
	if !assert.True(t, ok) {
		return
	}

	i.Bind("b", schema.Schema{Fields: schema.Fields{"id": {}}}, nil, DefaultConf)
	i.Bind("a", schema.Schema{Fields: schema.Fields{"ref": {
		Validator: &schema.Reference{Path: "c"},
	}}}, nil, DefaultConf)
	assert.Error(t, i.Compile())
}

func TestIndexGetResource(t *testing.T) {
	r := NewIndex()
	foo := r.Bind("foo", schema.Schema{}, nil, DefaultConf)
	foo.Bind("bar", "f", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	res, found := r.GetResource("foo", nil)
	assert.True(t, found)
	assert.Equal(t, "foo", res.name)
	assert.Equal(t, "", res.parentField)
	res, found = r.GetResource("foo.bar", nil)
	assert.True(t, found)
	assert.Equal(t, "bar", res.name)
	assert.Equal(t, "f", res.parentField)
	res, found = r.GetResource("foo.bar.baz", nil)
	assert.False(t, found)
	assert.Nil(t, res)
	res, found = r.GetResource("bar", nil)
	assert.False(t, found)
	assert.Nil(t, res)
	res, found = r.GetResource(".bar", foo)
	assert.True(t, found)
	assert.Equal(t, "bar", res.name)
	assert.Equal(t, "f", res.parentField)
	res, found = r.GetResource(".bar", nil)
	assert.False(t, found)
	assert.Nil(t, res)
}

func TestIndexGetResources(t *testing.T) {
	i := NewIndex()
	i.Bind("b", schema.Schema{}, nil, DefaultConf)
	i.Bind("a", schema.Schema{}, nil, DefaultConf)
	i.Bind("c", schema.Schema{}, nil, DefaultConf)
	if assert.Len(t, i.GetResources(), 3) {
		assert.Equal(t, "a", i.GetResources()[0].Name())
		assert.Equal(t, "b", i.GetResources()[1].Name())
		assert.Equal(t, "c", i.GetResources()[2].Name())
	}
}
