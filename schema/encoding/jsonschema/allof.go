package jsonschema

import (
	"errors"

	"github.com/rs/rest-layer/schema"
)

type allOfBuilder schema.AllOf

var (
	//ErrNoSchemaList is returned when trying to JSON Encode an empty
	//schema.AnyOf or schema.AllOf slice.
	ErrNoSchemaList = errors.New("at least one schema must be specified")
)

func (v allOfBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	if len(v) == 0 {
		return nil, ErrNoSchemaList
	}

	subSchemas := make([]map[string]interface{}, 0, len(v))

	for i := range v {
		b, err := ValidatorBuilder(v[i])
		if err != nil {
			return nil, err
		}
		schema, err := b.BuildJSONSchema()
		if err != nil {
			return nil, err
		}
		subSchemas = append(subSchemas, schema)
	}

	return map[string]interface{}{
		"allOf": subSchemas,
	}, nil

}
