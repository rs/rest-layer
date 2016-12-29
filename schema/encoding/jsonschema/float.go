package jsonschema

import (
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/rs/rest-layer/schema"
)

func encodeFloat(w io.Writer, v *schema.Float) error {
	ew := errWriter{w: w}

	ew.writeString(`"type": "number"`)
	if len(v.Allowed) > 0 {
		var allowed []string
		for _, value := range v.Allowed {
			allowed = append(allowed, strconv.FormatFloat(value, 'E', -1, 64))
		}
		ew.writeFormat(`, "enum": [%s]`, strings.Join(allowed, ","))
	}
	if ew.err == nil {
		ew.err = boundariesToJSONSchema(w, v.Boundaries)
	}
	return ew.err
}

// boundariesToJSONSchema writes JSON Schema keys and values based on b, prefixed by a comma and without curly braces,
// to w. The prefixing comma is only written if at least one key/value pair is also written.
func boundariesToJSONSchema(w io.Writer, b *schema.Boundaries) error {
	if b == nil {
		return nil
	}
	ew := errWriter{w: w}

	if !math.IsNaN(b.Min) && !math.IsInf(b.Min, -1) {
		ew.writeFormat(`, "minimum": %s`, strconv.FormatFloat(b.Min, 'E', -1, 64))
	}
	if !math.IsNaN(b.Max) && !math.IsInf(b.Max, 1) {
		ew.writeFormat(`, "maximum": %s`, strconv.FormatFloat(b.Max, 'E', -1, 64))
	}
	return ew.err
}
