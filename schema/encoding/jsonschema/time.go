package jsonschema

import (
	"io"

	"github.com/rs/rest-layer/schema"
)

func encodeTime(w io.Writer, v *schema.Time) error {
	ew := errWriter{w: w}
	ew.writeString(`"type": "string", "format": "date-time"`)
	return ew.err
}
