package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareEtag(t *testing.T) {
	assert.True(t, compareEtag(`abc`, `abc`))
	assert.True(t, compareEtag(`"abc"`, `abc`))
	assert.False(t, compareEtag(`'abc'`, `abc`))
	assert.False(t, compareEtag(`"abc`, `abc`))
	assert.False(t, compareEtag(``, `abc`))
	assert.False(t, compareEtag(`"cba"`, `abc`))
}
