package jsonschema

import (
	"io"

	"github.com/rs/rest-layer/schema"
)

func encodeObject(w io.Writer, v *schema.Object) error {
	ew := errWriter{w: w}
	if ew.err == nil && v.Schema != nil {
		ew.err = encodeSchema(w, v.Schema)
	}
	return ew.err
}
