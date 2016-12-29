package jsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/rest-layer/schema"
)

var (
	// ErrNotImplemented is returned when the JSON schema encoding logic for a schema.FieldValidator has not (yet)
	// been implemented.
	ErrNotImplemented = errors.New("not implemented")
)

// encodeSchema writes JSON Schema keys and values based on s, without the outer curly braces, to w.
func encodeSchema(w io.Writer, s *schema.Schema) (err error) {
	if s == nil {
		return
	}

	ew := errWriter{w: w}
	if s.Description != "" {
		ew.writeFormat(`"description": %q, `, s.Description)
	}
	ew.writeString(`"type": "object", `)
	ew.writeString(`"additionalProperties": false, `)
	ew.writeString(`"properties": {`)
	var required []string
	var notFirst bool
	for _, key := range sortedFieldNames(s.Fields) {
		field := s.Fields[key]
		if notFirst {
			ew.writeString(", ")
		}
		notFirst = true
		if field.Required {
			required = append(required, fmt.Sprintf("%q", key))
		}
		ew.err = encodeField(ew, key, field)
		if ew.err != nil {
			return ew.err
		}
	}
	ew.writeString("}")
	if s.MinLen > 0 {
		ew.writeFormat(`, "minProperties": %s`, strconv.FormatInt(int64(s.MinLen), 10))
	}
	if s.MaxLen > 0 {
		ew.writeFormat(`, "maxProperties": %s`, strconv.FormatInt(int64(s.MaxLen), 10))
	}

	if len(required) > 0 {
		ew.writeFormat(`, "required": [%s]`, strings.Join(required, ", "))
	}
	return ew.err
}

type fieldWriter struct {
	errWriter
	propertiesCount int
}

// comma optionally outputs a comma.  Invoke this when you're about to write a property.  Tracks how many have been
// written and emits if not the first.
func (fw *fieldWriter) comma() {
	if fw.propertiesCount > 0 {
		fw.writeString(",")
	}
	fw.propertiesCount++
}

func (fw *fieldWriter) resetPropertiesCount() {
	fw.propertiesCount = 0
}

// sortedFieldNames returns a list with all field names alphabetically sorted.
func sortedFieldNames(v schema.Fields) []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func encodeField(ew errWriter, key string, field schema.Field) error {
	fw := fieldWriter{ew, 0}
	fw.writeFormat("%q: {", key)
	if field.Description != "" {
		fw.comma()
		fw.writeFormat(`"description": %q`, field.Description)
	}
	if field.ReadOnly {
		fw.comma()
		fw.writeFormat(`"readOnly": %t`, field.ReadOnly)
	}
	if field.Validator != nil {
		// FIXME: This breaks if there are any Validators that may write nothing. E.g. a schema.Object with
		// Schema set to nil. A better solution should be found before adding support for custom validators.
		fw.comma()
		fw.err = encodeValidator(ew, field.Validator)
	}
	if field.Default != nil {
		b, err := json.Marshal(field.Default)
		if err != nil {
			return err
		}
		fw.comma()
		fw.writeString(`"default": `)
		fw.writeBytes(b)
	}
	fw.writeString("}")
	fw.resetPropertiesCount()
	return fw.err
}

// encodeValidator writes JSON Schema keys and values based on v, without the outer curly braces, to w. Note
// that not all FieldValidator types are supported at the moment.
func encodeValidator(w io.Writer, v schema.FieldValidator) (err error) {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case *schema.String:
		err = encodeString(w, t)
	case *schema.Integer:
		err = encodeInteger(w, t)
	case *schema.Float:
		err = encodeFloat(w, t)
	case *schema.Array:
		err = encodeArray(w, t)
	case *schema.Object:
		// FIXME: May break the JSON encoding atm. if t.Schema is nil.
		err = encodeObject(w, t)
	case *schema.Time:
		err = encodeTime(w, t)
	case *schema.Bool:
		err = encodeBool(w, t)
	default:
		return ErrNotImplemented
	}
	return err
}
