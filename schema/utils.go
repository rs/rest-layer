package schema

import (
	"strings"
	"time"
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

// isTime takes an interface as input, and returns a time.Time if the type is
// a string and is formatted with RFC3339
func isTime(t interface{}) (time.Time, bool) {
	var t2 time.Time
	var isTime bool
	if tStr, ok := t.(string); ok {
		t2, err := time.Parse(time.RFC3339, tStr)
		if err != nil {
			isTime = false
		} else {
			isTime = true
		}
		return t2, isTime
	}
	return t2, false
}

// getField gets the value of a given field by supporting sub-field path.
// A get on field.subfield is equivalent to payload["field"]["subfield].
func getField(payload map[string]interface{}, name string) interface{} {
	val, found := getFieldExist(payload, name)
	if !found {
		return nil
	}
	return val
}

func getFieldExist(payload map[string]interface{}, name string) (interface{}, bool) {
	// Split the name to get the current level name on first element and
	// the rest of the path as second element if dot notation is used
	// (i.e.: field.subfield.subsubfield -> field, subfield.subsubfield)
	path := strings.SplitN(name, ".", 2)
	if value, found := payload[path[0]]; found {
		if len(path) == 2 {
			if subPayload, ok := value.(map[string]interface{}); ok {
				// Check next level
				return getFieldExist(subPayload, path[1])
			}
			// The requested depth does not exist
			return nil, false
		}
		// Full path has been found
		return value, true
	}
	return nil, false
}
