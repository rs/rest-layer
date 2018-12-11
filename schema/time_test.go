package schema_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/rs/rest-layer/schema"
)

var testTimeFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
}

func TestTimeValidate(t *testing.T) {
	now := time.Now().Truncate(time.Minute).UTC()
	timeT := schema.Time{}
	err := timeT.Compile(nil)
	assert.NoError(t, err)
	for _, f := range testTimeFormats {
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
	timeT := schema.Time{TimeLayouts: []string{time.RFC3339, time.RFC822Z, time.RFC1123Z}}
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
	timeT := schema.Time{TimeLayouts: []string{time.RFC3339}}
	err := timeT.Compile(nil)
	assert.NoError(t, err)
	// expect an error
	for _, f := range testList {
		_, err := timeT.Validate(now.Format(f))
		assert.EqualError(t, err, "not a time")
	}
}

func TestTimeLess(t *testing.T) {
	low, _ := time.Parse(time.RFC3339, "2018-11-18T17:15:16Z")
	high, _ := time.Parse(time.RFC3339, "2018-11-19T17:15:16Z")
	cases := []struct {
		name         string
		value, other interface{}
		expected     bool
	}{
		{`Time.Less(time.Time-low,time.Time-high)`, low, high, true},
		{`Time.Less(time.Time-low,time.Time-low)`, low, low, false},
		{`Time.Less(time.Time-high,time.Time-low)`, high, low, false},
		{`Time.Less(time.Time,string)`, low, "2.0", false},
	}
	lessFunc := schema.Time{}.LessFunc()
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := lessFunc(tt.value, tt.other)
			if got != tt.expected {
				t.Errorf("output for `%v`\ngot:  %v\nwant: %v", tt.name, got, tt.expected)
			}
		})
	}
}
