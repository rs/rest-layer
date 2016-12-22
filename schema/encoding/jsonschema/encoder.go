package jsonschema

import (
	"fmt"
	"io"

	"github.com/rs/rest-layer/schema"
)

// Encoder writes the JSON Schema representation of a schema.Schema to an output stream. Note that only a sub-set of the
// FieldValidator types in the schema package is supported at the moment. Custom validators are also not yet handled.
// Attempting to encode a schema containing such fields will result in a ErrNotImplemented error.
type Encoder struct {
	io.Writer
}

// NewEncoder returns a new JSONSchema Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

// Encode writes the JSON Schema representation of s to the stream, followed by a newline character.
func (e *Encoder) Encode(s *schema.Schema) error {
	ew := errWriter{w: e.Writer}
	ew.writeString("{")
	if ew.err == nil {
		ew.err = encodeSchema(e.Writer, s)
	}
	ew.writeString("}\n")
	return ew.err
}

// Wrap IO writer so we can consolidate error handling in a single place. Also track properties written so we know when
// to emit a separator.
type errWriter struct {
	w   io.Writer // writer instance
	err error     // track errors
}

// Write ensures compatibility with the io.Writer interface.
func (ew errWriter) Write(p []byte) (int, error) {
	if ew.err != nil {
		return 0, ew.err
	}
	return ew.w.Write(p)
}

func (ew errWriter) writeFormat(format string, a ...interface{}) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintf(ew.w, format, a...)
}

func (ew errWriter) writeString(s string) {
	if ew.err != nil {
		return
	}
	_, ew.err = ew.w.Write([]byte(s))
}

func (ew errWriter) writeBytes(b []byte) {
	if ew.err != nil {
		return
	}
	_, ew.err = ew.w.Write(b)
}
