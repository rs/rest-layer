package jsonschema_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/encoding/jsonschema"
)

func ExampleEncoder() {
	s := schema.Schema{
		Fields: schema.Fields{
			"foo": schema.Field{
				Required: true,
				// NOTE: Min is currently encoded as '0E+00', not '0'.
				Validator: &schema.Float{Boundaries: &schema.Boundaries{Min: 0, Max: math.Inf(1)}},
			},
			"bar": schema.Field{
				Validator: &schema.Integer{},
			},
			"baz": schema.Field{
				ReadOnly:  true,
				Validator: &schema.String{MaxLen: 42},
			},
			"foobar": schema.Field{},
		},
	}
	b := new(bytes.Buffer)
	enc := jsonschema.NewEncoder(b)
	enc.Encode(&s)
	b2 := new(bytes.Buffer)
	json.Indent(b2, b.Bytes(), "", "| ")
	fmt.Println(b2)
	// Output: {
	// | "type": "object",
	// | "additionalProperties": false,
	// | "properties": {
	// | | "bar": {
	// | | | "type": "integer"
	// | | },
	// | | "baz": {
	// | | | "readOnly": true,
	// | | | "type": "string",
	// | | | "maxLength": 42
	// | | },
	// | | "foo": {
	// | | | "type": "number",
	// | | | "minimum": 0E+00
	// | | },
	// | | "foobar": {}
	// | },
	// | "required": [
	// | | "foo"
	// | ]
	// }
}
