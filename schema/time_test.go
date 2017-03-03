package schema

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeValidate(t *testing.T) {
	now := time.Now().Truncate(time.Minute).UTC()
	timeT := Time{}
	err := timeT.Compile(nil)
	assert.NoError(t, err)
	for _, f := range formats {
		v, err := timeT.Validate(now.Format(f))
		assert.NoError(t, err)
		if assert.IsType(t, v, now) {
			assert.True(t, now.Equal(v.(time.Time)), f)
		}
	}
	v, err := timeT.Validate(now)
	assert.NoError(t, err)
	assert.Equal(t, now, v)
	v, err = timeT.Validate("invalid date")
	assert.EqualError(t, err, "not a time")
	assert.Nil(t, v)
}

func TestTimeSpecificLayoutList(t *testing.T) {
	now := time.Now().Truncate(time.Minute).UTC()
	// list to test for
	testList := []string{time.RFC1123Z, time.RFC822Z, time.RFC3339}
	// test for same list in reverse
	timeT := Time{TimeLayouts: []string{time.RFC3339, time.RFC822Z, time.RFC1123Z}}
	err := timeT.Compile(nil)
	assert.NoError(t, err)
	// expect no errors
	for _, f := range testList {
		_, err := timeT.Validate(now.Format(f))
		assert.NoError(t, err)
	}
}

func TestTimeForTimeLayoutFailure(t *testing.T) {
	now := time.Now().Truncate(time.Minute).UTC()
	// test for ANSIC time
	testList := []string{time.ANSIC}
	// configure for RFC3339 time
	timeT := Time{TimeLayouts: []string{time.RFC3339}}
	err := timeT.Compile(nil)
	assert.NoError(t, err)
	// expect an error
	for _, f := range testList {
		_, err := timeT.Validate(now.Format(f))
		assert.EqualError(t, err, "not a time")
	}
}
