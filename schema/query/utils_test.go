package query

import (
	"testing"
)

func TestIsNumber(t *testing.T) {
	cases := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{"1", 1, true},
		{"int8(1)", int8(1), true},
		{"int16(1)", int16(1), true},
		{"int32(1)", int32(1), true},
		{"int64(1)", int64(1), true},
		{"uint(1)", uint(1), true},
		{"uint8(1)", uint8(1), true},
		{"uint16(1)", uint16(1), true},
		{"uint32(1)", uint32(1), true},
		{"uint64(1)", uint64(1), true},
		{"float32(1)", float32(1), true},
		{"float64(1)", float64(1), true},
		{`"1"`, "1", false},
	}
	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			_, ok := isNumber(tc.input)
			if ok != tc.want {
				t.Errorf("isNumber = %v, wanted %v", ok, tc.want)
			}
		})
	}
}

func TestGetField(t *testing.T) {
	cases := []struct {
		name      string
		payload   map[string]interface{}
		fieldName string
		want      string
	}{
		{"foo", map[string]interface{}{"foo": "bar"}, "foo", "bar"},
		{"foo.bar", map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}, "foo.bar", "baz"},
	}
	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			if res := getField(tc.payload, tc.fieldName); res != tc.want {
				t.Errorf("field = %v, wanted %v", res, tc.want)
			}
		})
	}
}
