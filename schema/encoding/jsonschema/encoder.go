package jsonschema

import (
	"encoding/json"
	"io"

	"github.com/rs/rest-layer/schema"
)

// Encoder writes the JSON Schema representation of a schema.Schema to an output
// stream. Note that only a sub-set of the FieldValidator types in the schema
// package is supported at the moment. Custom validators are also not yet
// handled. Attempting to encode a schema containing such fields will result in
// a ErrNotImplemented error.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new JSONSchema Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the JSON Schema representation of s to the stream, followed by
// a newline character.
func (e *Encoder) Encode(s *schema.Schema) error {
	m := make(map[string]interface{})
	if err := addSchemaProperties(m, s); err != nil {
		return err
	}
	enc := json.NewEncoder(e.w)
	return enc.Encode(m)

}

// The Builder interface should be implemented by custom schema.FieldValidator
// implementations to allow JSON Schema serialization.
type Builder interface {
	// BuildJSONSchema should return a map containing JSON Schema Draft 4
	// properties that can be set based on FieldValidator data. Application
	// specific properties can be added as well, but should not conflict with
	// any legal JSON Schema keys.
	BuildJSONSchema() (map[string]interface{}, error)
}

// builderFunc is an adapter that allows pure functions to implement the Builder
// interface.
type builderFunc func() (map[string]interface{}, error)

func (f builderFunc) BuildJSONSchema() (map[string]interface{}, error) {
	return f()
}
