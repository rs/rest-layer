package jsonschema

import "github.com/rs/rest-layer/schema"

type arrayBuilder schema.Array

func (v arrayBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	m := map[string]interface{}{
		"type": "array",
	}
	if v.MinLen > 0 {
		m["minItems"] = v.MinLen
	}
	if v.MaxLen > 0 {
		m["maxItems"] = v.MaxLen
	}
	if v.ValuesValidator != nil {
		builder, err := ValidatorBuilder(v.ValuesValidator)
		if err != nil {
			return nil, err
		}
		items, err := builder.BuildJSONSchema()
		if err != nil {
			return nil, err
		}
		m["items"] = items
	}
	return m, nil
}
