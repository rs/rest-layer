package jsonschema

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/rs/rest-layer/schema"
)

func encodeString(w io.Writer, v *schema.String) error {
	ew := errWriter{w: w}
	ew.writeString(`"type": "string"`)
	if v.Regexp != "" {
		ew.writeFormat(`, "pattern": %q`, v.Regexp)
	}
	if len(v.Allowed) > 0 {
		var allowed []string
		for _, value := range v.Allowed {
			allowed = append(allowed, fmt.Sprintf("%q", value))
		}
		ew.writeFormat(`, "enum": [%s]`, strings.Join(allowed, ", "))
	}
	if v.MinLen > 0 {
		ew.writeFormat(`, "minLength": %s`, strconv.FormatInt(int64(v.MinLen), 10))
	}
	if v.MaxLen > 0 {
		ew.writeFormat(`, "maxLength": %s`, strconv.FormatInt(int64(v.MaxLen), 10))
	}
	return ew.err
}
