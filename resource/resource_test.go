package resource

import (
	"io/ioutil"
	"log"
	"net/url"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestResourceBind(t *testing.T) {
	barSchema := schema.Schema{Fields: schema.Fields{"foo": {}}}
	i := NewIndex()
	foo := i.Bind("foo", schema.Schema{}, nil, DefaultConf)
	bar := foo.Bind("bar", "foo", barSchema, nil, DefaultConf)
	assert.Equal(t, "bar", bar.Name())
	assert.Equal(t, "foo.bar", bar.Path())
	assert.Equal(t, "foo", bar.ParentField())
	assert.Len(t, foo.GetResources(), 1)
	assert.Len(t, bar.GetResources(), 0)
	assert.Equal(t, schema.Schema{Fields: schema.Fields{"foo": {}}}, bar.Schema())
	assert.Equal(t, validatorFallback{
		Validator: schema.Schema{},
		fallback: schema.Schema{Fields: schema.Fields{
			"bar": {
				ReadOnly: true,
				Validator: &schema.Connection{
					Path:      ".bar",
					Field:     "foo",
					Validator: bar.validator,
				},
				Params: schema.Params{
					"skip": schema.Param{
						Description: "The number of items to skip",
						Validator: schema.Integer{
							Boundaries: &schema.Boundaries{Min: 0},
						},
					},
					"page": schema.Param{
						Description: "The page number",
						Validator: schema.Integer{
							Boundaries: &schema.Boundaries{Min: 1, Max: 1000},
						},
					},
					"limit": schema.Param{
						Description: "The number of items to return per page",
						Validator: schema.Integer{
							Boundaries: &schema.Boundaries{Min: 0, Max: 1000},
						},
					},
					"sort": schema.Param{
						Description: "The field(s) to sort on",
						Validator:   schema.String{},
					},
					"filter": schema.Param{
						Description: "The filter query",
						Validator:   schema.String{},
					},
				},
			},
		}},
	}, foo.Validator())
	assert.Equal(t, DefaultConf, bar.Conf())
}

func TestResourceAlias(t *testing.T) {
	i := NewIndex()
	foo := i.Bind("foo", schema.Schema{}, nil, DefaultConf)
	foo.Alias("foo", url.Values{"bar": []string{"baz"}})
	a, found := foo.GetAlias("foo")
	assert.True(t, found)
	assert.Equal(t, url.Values{"bar": []string{"baz"}}, a)
	_, found = foo.GetAlias("bar")
	assert.False(t, found)
	assert.Equal(t, []string{"foo"}, foo.GetAliases())
}

func TestSubResources(t *testing.T) {
	sr := subResources{}
	sr.add(&Resource{name: "b"})
	sr.add(&Resource{name: "a"})
	sr.add(&Resource{name: "c"})
	sr.add(&Resource{name: "e"})
	if assert.Len(t, sr, 4) {
		assert.Equal(t, "a", sr[0].name)
		assert.Equal(t, "b", sr[1].name)
		assert.Equal(t, "c", sr[2].name)
		assert.Equal(t, "e", sr[3].name)
	}
	assert.Equal(t, "a", sr.get("a").name)
	assert.Equal(t, "b", sr.get("b").name)
	assert.Equal(t, "c", sr.get("c").name)
	assert.Equal(t, "e", sr.get("e").name)
	assert.Nil(t, sr.get("0"))
	assert.Nil(t, sr.get("d"))
	assert.Nil(t, sr.get("f"))
}

func TestResourceBindDupViaAlias(t *testing.T) {
	r := newResource("name", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	r.Alias("foo", url.Values{})
	log.SetOutput(ioutil.Discard)
	assert.Panics(t, func() {
		r.Bind("foo", "f", schema.Schema{}, nil, DefaultConf)
	})
}

func TestResourceBindOnMissingField(t *testing.T) {
	r := newResource("name", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	log.SetOutput(ioutil.Discard)
	assert.Panics(t, func() {
		r.Bind("foo", "m", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	})
}

func TestResourceValidatorFallback(t *testing.T) {
	vf := validatorFallback{
		Validator: schema.Schema{Fields: schema.Fields{"foo": {}}},
		fallback:  schema.Schema{Fields: schema.Fields{"bar": {}}},
	}
	assert.NotNil(t, vf.GetField("foo"))
	assert.NotNil(t, vf.GetField("bar"))
	assert.Nil(t, vf.GetField("baz"))
}

func TestResourceConnection(t *testing.T) {
	c := schema.Connection{}
	v, err := c.Validate("foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", v)
}

func TestResourceUse(t *testing.T) {
	i := NewIndex()
	r := i.Bind("foo", schema.Schema{}, nil, DefaultConf)
	r.Use(FindEventHandlerFunc(nil))
	assert.Len(t, r.hooks.onFindH, 1)

	err := r.Use("non handler")
	assert.EqualError(t, err, "does not implement any event handler interface")
}
