package encoding

import (
	"bytes"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/encoding/jsonschema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJSONSchemaEncoder(t *testing.T) {
	b := new(bytes.Buffer)
	encoder := jsonschema.NewEncoder(b)

	s := &schema.Schema{
		Fields: schema.Fields{
			"name": schema.Field{
				Validator: &schema.String{},
			},
		},
	}
	assert.Nil(t, encoder.Encode(s))

}
