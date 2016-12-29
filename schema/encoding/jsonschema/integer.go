package jsonschema

import (
	"io"
	"strconv"
	"strings"

	"github.com/rs/rest-layer/schema"
)

func encodeInteger(w io.Writer, v *schema.Integer) error {
	ew := errWriter{w: w}

	ew.writeString(`"type": "integer"`)
	if len(v.Allowed) > 0 {
		var allowed []string
		for _, value := range v.Allowed {
			allowed = append(allowed, strconv.FormatInt(int64(value), 10))
		}
		ew.writeFormat(`, "enum": [%s]`, strings.Join(allowed, ","))
	}
	if ew.err == nil {
		ew.err = boundariesToJSONSchema(w, v.Boundaries)
	}
	return ew.err
}
