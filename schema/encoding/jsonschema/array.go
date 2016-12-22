package jsonschema

import (
	"io"
	"strconv"

	"github.com/rs/rest-layer/schema"
)

func encodeArray(w io.Writer, v *schema.Array) error {
	ew := errWriter{w: w}

	ew.writeString(`"type": "array"`)
	if v.MinLen > 0 {
		ew.writeFormat(`, "minItems": %s`, strconv.FormatInt(int64(v.MinLen), 10))
	}
	if v.MaxLen > 0 {
		ew.writeFormat(`, "maxItems": %s`, strconv.FormatInt(int64(v.MaxLen), 10))
	}
	if v.ValuesValidator != nil {
		ew.writeString(`, "items": {`)
		if ew.err == nil {
			ew.err = encodeValidator(w, v.ValuesValidator)
		}
		ew.writeString("}")
	}
	return ew.err
}
