package jsonschema

import "github.com/rs/rest-layer/schema"

type anyOfBuilder schema.AnyOf

func (v anyOfBuilder) BuildJSONSchema() (map[string]interface{}, error) {
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
		"anyOf": subSchemas,
	}, nil

}
