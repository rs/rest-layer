package schema

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	// Now is a field hook handler that returns the current time, to be used in
	// schema with OnInit and OnUpdate
	Now = func(value interface{}) interface{} {
		return time.Now()
	}

	// NewID is a field hook handler that generates a new unique id if none exist,
	// to be used in schema with OnInit
	NewID = func(value interface{}) interface{} {
		if value == nil {
			r := make([]byte, 128)
			rand.Read(r)
			value = fmt.Sprintf("%x", md5.Sum(r))
		}
		return value
	}

	// IDField is a common schema field configuration that generate an UUID for new item id.
	IDField = Field{
		Required:   true,
		ReadOnly:   true,
		OnInit:     &NewID,
		Filterable: true,
		Sortable:   true,
		Validator: &String{
			Regexp: "^[0-9a-f]{32}$",
		},
	}

	// CreatedField is a common schema field configuration for "created" fields. It stores
	// the creation date of the item.
	CreatedField = Field{
		Required:  true,
		ReadOnly:  true,
		OnInit:    &Now,
		Sortable:  true,
		Validator: &Time{},
	}

	// UpdatedField is a common schema field configuration for "updated" fields. It stores
	// the current date each time the item is modified.
	UpdatedField = Field{
		Required:  true,
		ReadOnly:  true,
		OnInit:    &Now,
		OnUpdate:  &Now,
		Sortable:  true,
		Validator: &Time{},
	}
)

// isNumber takes an interface as input, and returns a float64 if the type is
// compatible (int* or float*)
func isNumber(n interface{}) (float64, bool) {
	switch n := n.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

// isIn returns true if on the the elements in exp is equal to value.
// The exp argument may be an item or a list of item to match.
func isIn(exp interface{}, value interface{}) bool {
	values, ok := exp.([]interface{})
	if !ok {
		values = []interface{}{exp}
	}
	for _, v := range values {
		if reflect.DeepEqual(v, value) {
			return true
		}
	}
	return false
}

// getField gets the value of a given field by supporting sub-field path.
// A get on field.subfield is equivalent to payload["field"]["subfield].
func getField(payload map[string]interface{}, name string) interface{} {
	// Split the name to get the current level name on first element and
	// the rest of the path as second element if dot notation is used
	// (i.e.: field.subfield.subsubfield -> field, subfield.subsubfield)
	path := strings.SplitN(name, ".", 2)
	if value, found := payload[path[0]]; found {
		if len(path) == 2 {
			if subPayload, ok := value.(map[string]interface{}); ok {
				// Check next level
				return getField(subPayload, path[1])
			}
			// The requested depth does not exist
			return nil
		}
		// Full path has been found
		return value
	}
	return nil
}
