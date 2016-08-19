package jsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/rest-layer/schema"
	"io"
	"strconv"
	"strings"
)

var (
	// ErrNotImplemented means schema.FieldValidator type is not implemented
	ErrNotImplemented = errors.New("Schema not Implemented")
)

type errWriter struct {
	w   io.Writer
	err error
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

func (ew errWriter) write(b []byte) {
	if ew.err != nil {
		return
	}
	_, ew.err = ew.w.Write(b)
}

// ValidatorToJSONSchema takes a validator and renders to JSON
func validatorToJSONSchema(w io.Writer, v schema.FieldValidator) (err error) {
	if v == nil {
		return nil
	}
	ew := errWriter{w: w}
	switch t := v.(type) {
	case *schema.String:
		ew.writeString(`"type": "string"`)
		if t.Regexp != "" {
			ew.writeFormat(`, "pattern": %q`, t.Regexp)
		}
		if len(t.Allowed) > 0 {
			var allowed []string
			for _, value := range t.Allowed {
				allowed = append(allowed, fmt.Sprintf("%q", value))
			}
			ew.writeFormat(`, "enum": [%s]`, strings.Join(allowed, ", "))
		}
		if t.MinLen > 0 {
			ew.writeFormat(`, "minLength": %s`, strconv.FormatInt(int64(t.MinLen), 10))
		}
		if t.MaxLen > 0 {
			ew.writeFormat(`, "maxLength": %s`, strconv.FormatInt(int64(t.MaxLen), 10))
		}
	case *schema.Integer:
		ew.writeString(`"type": "integer"`)

		if len(t.Allowed) > 0 {
			var allowed []string
			for _, value := range t.Allowed {
				allowed = append(allowed, strconv.FormatInt(int64(value), 10))
			}
			ew.writeFormat(`, "enum": [%s]`, strings.Join(allowed, ","))
		}
		if t.Boundaries != nil {
			ew.writeFormat(`, "minimum": %s, "maximum": %s`,
				strconv.FormatFloat(t.Boundaries.Min, 'E', -1, 64),
				strconv.FormatFloat(t.Boundaries.Max, 'E', -1, 64))
		}
	case *schema.Float:
		ew.writeString(`"type": "number"`)
		if len(t.Allowed) > 0 {
			var allowed []string
			for _, value := range t.Allowed {
				allowed = append(allowed, strconv.FormatFloat(value, 'E', -1, 64))
			}
			ew.writeFormat(`, "enum": [%s]`, strings.Join(allowed, ","))
		}
		if t.Boundaries != nil {
			ew.writeFormat(`, "minimum": %s, "maximum": %s`,
				strconv.FormatFloat(t.Boundaries.Min, 'E', -1, 64),
				strconv.FormatFloat(t.Boundaries.Max, 'E', -1, 64))
		}

	case *schema.Array:
		ew.writeString(`"type": "array"`)
		if t.ValuesValidator != nil {
			ew.writeString(`, "items": `)
			if ew.err == nil {
				ew.err = validatorToJSONSchema(w, t.ValuesValidator)
			}
		}
	case *schema.Object:
		if ew.err == nil {
			ew.err = schemaToJSONSchema(w, t.Schema)
		}
	case *schema.Time:
		ew.writeString(`"type": "string", "format": "date-time"`)
	case *schema.Bool:
		ew.writeString(`"type": "boolean"`)
	default:
		return ErrNotImplemented
	}
	return ew.err
}

// SchemaToJSONSchema helper
func schemaToJSONSchema(w io.Writer, s *schema.Schema) (err error) {
	ew := errWriter{w: w}
	ew.writeString("{")
	if s.Description != "" {
		ew.writeFormat(`"title": %q, `, s.Description)
	}
	ew.writeString(`"type": "object", `)
	ew.writeString(`"properties": {`)
	var required []string
	var notFirst bool
	for key, field := range s.Fields {
		if notFirst {
			ew.writeString(", ")
		}
		notFirst = true
		ew.writeFormat("%q: {", key)
		if field.Description != "" {
			ew.writeFormat(`"description": %q, `, field.Description)
		}
		if field.Required {
			required = append(required, fmt.Sprintf("%q", key))
		}
		if field.ReadOnly {
			ew.writeFormat(`"readOnly": %t, `, field.ReadOnly)
		}
		if ew.err == nil {
			ew.err = validatorToJSONSchema(w, field.Validator)
		}
		if field.Default != nil {
			b, err := json.Marshal(field.Default)
			if err != nil {
				return err
			}
			ew.writeString(`, "default": `)
			ew.write(b)
		}
		ew.writeString("}")
		if ew.err != nil {
			break
		}
	}
	ew.writeFormat(`, "required": [%s]`, strings.Join(required, ", "))
	ew.writeString("}}")
	return ew.err
}

// Encoder encodes schema.Schema into a JSONSchema
type Encoder struct {
	io.Writer
}

// NewEncoder returns a new JSONSchema Encoder.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

// Encode is take schema and writes to Writer
func (e *Encoder) Encode(s *schema.Schema) error {
	return schemaToJSONSchema(e.Writer, s)
}
