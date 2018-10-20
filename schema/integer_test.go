package schema

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegerQueryValidator(t *testing.T) {
	cases := []struct {
		name          string
		field         Integer
		input, expect interface{}
		err           error
	}{
		{`Integer.ValidateQuery(int)`, Integer{}, 1, 1, nil},
		{`Integer.ValidateQuery(float64)`, Integer{}, 1.1, nil, errors.New("not an integer")},
		{`Integer.ValidateQuery(string)`, Integer{}, "1", nil, errors.New("not an integer")},
		{"Integer.ValidateQuery(int)-out of range above", Integer{Boundaries: &Boundaries{Min: 0, Max: 2}}, 3, 3, nil},
		{"Integer.ValidateQuery(int)-in range", Integer{Boundaries: &Boundaries{Min: 0, Max: 2}}, 1, 1, nil},
		{"Integer.ValidateQuery(int)-out of range below", Integer{Boundaries: &Boundaries{Min: 0, Max: 2}}, -1, -1, nil},
	}
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := (tt.field).ValidateQuery(tt.input)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("unexpected error for `%v`\ngot:  %v\nwant: %v", tt.input, err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("invalid output for `%v`:\ngot:  %#v\nwant: %#v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestIntegerValidator(t *testing.T) {
	s, err := Integer{}.Validate(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, s)
	s, err = Integer{}.Validate(1.1)
	assert.EqualError(t, err, "not an integer")
	assert.Nil(t, s)
	s, err = Integer{}.Validate("1")
	assert.EqualError(t, err, "not an integer")
	assert.Nil(t, s)
	s, err = Integer{Boundaries: &Boundaries{Min: 0, Max: 2}}.Validate(3)
	assert.EqualError(t, err, "is greater than 2")
	assert.Nil(t, s)
	s, err = Integer{Boundaries: &Boundaries{Min: 0, Max: 2}}.Validate(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, s)
	s, err = Integer{Boundaries: &Boundaries{Min: 2, Max: 10}}.Validate(1)
	assert.EqualError(t, err, "is lower than 2")
	assert.Nil(t, s)
	s, err = Integer{Boundaries: &Boundaries{Min: 2, Max: 10}}.Validate(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, s)
	s, err = Integer{Boundaries: &Boundaries{Min: math.Inf(-1), Max: 10}}.Validate(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, s)
	s, err = Integer{Boundaries: &Boundaries{Min: math.NaN(), Max: 10}}.Validate(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, s)
	s, err = Integer{Boundaries: &Boundaries{Min: 2, Max: math.Inf(1)}}.Validate(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, s)
	s, err = Integer{Boundaries: &Boundaries{Min: 2, Max: math.NaN()}}.Validate(3)
	assert.NoError(t, err)
	assert.Equal(t, 3, s)
	s, err = Integer{Boundaries: &Boundaries{}}.Validate(1)
	assert.EqualError(t, err, "is greater than 0")
	assert.Nil(t, s)
	s, err = Integer{Boundaries: &Boundaries{}}.Validate(-1)
	assert.EqualError(t, err, "is lower than 0")
	assert.Nil(t, s)
	s, err = Integer{Allowed: []int{1, 2, 3}}.Validate(4)
	assert.EqualError(t, err, "not one of the allowed values")
	assert.Nil(t, s)
	s, err = Integer{Allowed: []int{1, 2, 3}}.Validate(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, s)
}

func TestIntegerParse(t *testing.T) {
	cases := []struct {
		name          string
		input, expect interface{}
		err           error
	}{
		{`Integer.parse(int)`, 1, 1, nil},
		{`Integer.parse(1.0)`, 1.0, 1, nil},
		{`Integer.parse(1.1)`, 1.1, nil, errors.New("not an integer")},
		{`Integer.parse(string)`, "1", nil, errors.New("not an integer")},
	}
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Integer{}.parse(tt.input)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("unexpected error for `%v`\ngot:  %v\nwant: %v", tt.input, err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("invalid output for `%v`:\ngot:  %#v\nwant: %#v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestIntegerGet(t *testing.T) {
	cases := []struct {
		name          string
		input, expect interface{}
		err           error
	}{
		{`Integer.get(int)`, 1, 1, nil},
		{`Integer.get(float64)`, 1.0, 0, errors.New("not an integer")},
	}
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Integer{}.get(tt.input)
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("unexpected error for `%v`\ngot:  %v\nwant: %v", tt.input, err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("invalid output for `%v`:\ngot:  %#v\nwant: %#v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestIntegerLesser(t *testing.T) {
	cases := []struct {
		name         string
		value, other interface{}
		expected     bool
	}{
		{`Integer.Less(1,2)`, 1, 2, true},
		{`Integer.Less(1,1)`, 1, 1, false},
		{`Integer.Less(2,1)`, 2, 1, false},
		{`Integer.Less(1,"2")`, 1, "2", false},
	}
	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Integer{}.Less(tt.value, tt.other)
			if got != tt.expected {
				t.Errorf("output for `%v`\ngot:  %v\nwant: %v", tt.name, got, tt.expected)
			}
		})
	}
}
