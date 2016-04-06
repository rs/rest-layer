package resource

import (
	"io/ioutil"
	"log"
	"net/url"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

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
	r := new("name", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	r.Alias("foo", url.Values{})
	log.SetOutput(ioutil.Discard)
	assert.Panics(t, func() {
		r.Bind("foo", "f", schema.Schema{}, nil, DefaultConf)
	})
}

func TestResourceBindOnMissingField(t *testing.T) {
	r := new("name", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	log.SetOutput(ioutil.Discard)
	assert.Panics(t, func() {
		r.Bind("foo", "m", schema.Schema{Fields: schema.Fields{"f": {}}}, nil, DefaultConf)
	})
}
