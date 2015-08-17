package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLValidator(t *testing.T) {
	u, err := URL{}.Validate("http://foo.com/bar")
	assert.NoError(t, err)
	assert.Equal(t, "http://foo.com/bar", u)
	u, err = URL{AllowRelative: true, AllowLocale: true, AllowNonHTTP: true}.Validate("/bar")
	assert.NoError(t, err)
	assert.Equal(t, "/bar", u)
	u, err = URL{}.Validate("/bar")
	assert.EqualError(t, err, "is relative URL")
	assert.Nil(t, u)
	u, err = URL{}.Validate("http://localhost/bar")
	assert.EqualError(t, err, "invalid domain")
	assert.Nil(t, u)
	u, err = URL{}.Validate("ftp://foo.com/bar")
	assert.EqualError(t, err, "invalid scheme")
	assert.Nil(t, u)
	u, err = URL{}.Validate("HTTP://foo.com/bar")
	assert.NoError(t, err)
	assert.Equal(t, "http://foo.com/bar", u)
	u, err = URL{}.Validate(":foo")
	assert.EqualError(t, err, "invalid URL: parse :foo: missing protocol scheme")
	assert.Nil(t, u)
	u, err = URL{}.Validate(1)
	assert.EqualError(t, err, "invalid type")
	assert.Nil(t, u)
	u, err = URL{AllowedSchemes: []string{"foo"}}.Validate("foo://foo.com/bar")
	assert.NoError(t, err)
	assert.Equal(t, "foo://foo.com/bar", u)
	u, err = URL{AllowedSchemes: []string{"foo"}}.Validate("http://foo.com/bar")
	assert.EqualError(t, err, "invalid scheme")
	assert.Nil(t, u)
}
