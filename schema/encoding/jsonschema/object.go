package jsonschema

import (
	"fmt"

	"github.com/rs/rest-layer/schema"
)

type objectBuilder schema.Object

var (
	//ErrNoSchema is returned when trying to JSON Encode a schema.Object with the Schema property set to nil.
	ErrNoSchema = fmt.Errorf("no schema defined for object")
)

func (v objectBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	if v.Schema == nil {
		return nil, ErrNoSchema
	}

	m := map[string]interface{}{}
	err := addSchemaProperties(m, v.Schema)
	if err != nil {
		return nil, err
	}
	return m, nil

}
