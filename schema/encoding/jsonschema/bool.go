package jsonschema

import (
	"io"

	"github.com/rs/rest-layer/schema"
)

func encodeBool(w io.Writer, v *schema.Bool) error {
	ew := errWriter{w: w}
	ew.writeString(`"type": "boolean"`)
	return ew.err
}
