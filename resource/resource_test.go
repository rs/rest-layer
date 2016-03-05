package resource

import (
	"io/ioutil"
	"log"
	"net/url"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

func TestResourceBindDupViaAlias(t *testing.T) {
	r := new("name", schema.Schema{Fields: schema.Fields{"f": schema.Field{}}}, nil, DefaultConf)
	r.Alias("foo", url.Values{})
	log.SetOutput(ioutil.Discard)
	assert.Panics(t, func() {
		r.Bind("foo", "f", schema.Schema{}, nil, DefaultConf)
	})
}

func TestResourceBindOnMissingField(t *testing.T) {
	r := new("name", schema.Schema{Fields: schema.Fields{"f": schema.Field{}}}, nil, DefaultConf)
	log.SetOutput(ioutil.Discard)
	assert.Panics(t, func() {
		r.Bind("foo", "m", schema.Schema{Fields: schema.Fields{"f": schema.Field{}}}, nil, DefaultConf)
	})
}
