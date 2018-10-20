package schema

import (
	"errors"
	"reflect"
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

func TestTimeParse(t *testing.T) {
	now := time.Now().Truncate(time.Minute).UTC()
	timeT := Time{}
	err := timeT.Compile(nil)
	if err != nil {
		t.Fail()
	}
	cases := []struct {
		name   string
		input  string
		expect interface{}
		err    error
	}{
		{`Time.parse(string)-valid`, now.Format(time.RFC3339), now, nil},
		{`Time.parse(string)-invalid`, "invalid", nil, errors.New("not a time")},
	}
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := timeT.parse(tt.input)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("unexpected error for `%v`\ngot:  %v\nwant: %v", tt.input, err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("invalid output for `%v`:\ngot:  %#v\nwant: %#v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestTimeGet(t *testing.T) {
	now := time.Now().Truncate(time.Minute).UTC()
	timeT := Time{}
	err := timeT.Compile(nil)
	if err != nil {
		t.Fail()
	}
	cases := []struct {
		name          string
		input, expect interface{}
		err           error
	}{
		{`Time.get(time.Time)`, now, now, nil},
		{`Time.get(RFC3339-string)`, now.Format(time.RFC3339), time.Time{}, errors.New("not a time")},
		{`Time.get(string)`, "invalid", time.Time{}, errors.New("not a time")},
	}
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := timeT.get(tt.input)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("unexpected error for `%v`\ngot:  %v\nwant: %v", tt.input, err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("invalid output for `%v`:\ngot:  %#v\nwant: %#v", tt.input, got, tt.expect)
			}
		})
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
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Time{}.Less(tt.value, tt.other)
			if got != tt.expected {
				t.Errorf("output for `%v`\ngot:  %v\nwant: %v", tt.name, got, tt.expected)
			}
		})
	}
}
