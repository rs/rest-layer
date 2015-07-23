package rest

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Etag computes an etag based on containt of the payload
func genEtag(payload map[string]interface{}) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(b)), nil
}

// getField gets the value of a given field by supporting sub-field path.
// A get on field.subfield is equivalent to payload["field"]["subfield].
func getField(name string, payload map[string]interface{}) interface{} {
	// Split the name to get the current level name on first element and
	// the rest of the path as second element if dot notation is used
	// (i.e.: field.subfield.subsubfield -> field, subfield.subsubfield)
	path := strings.SplitN(name, ".", 2)
	if value, found := payload[path[0]]; found {
		if len(path) == 2 {
			if subPayload, ok := value.(map[string]interface{}); ok {
				// Check next level
				return getField(path[1], subPayload)
			}
			// The requested depth does not exist
			return nil
		}
		// Full path has been found
		return value
	}
	return nil
}

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

// isIn returns true if on the the elements in exp is equal value.
// The exp argument may be an item or a list of item to match.
func isIn(exp interface{}, value interface{}) bool {
	// Handle both {$in: val} and {$in: [val1, va12]}
	conds, ok := exp.([]interface{})
	if !ok {
		conds = []interface{}{exp}
	}
	for _, cond := range conds {
		if reflect.DeepEqual(cond, value) {
			return true
		}
	}
	return false
}
