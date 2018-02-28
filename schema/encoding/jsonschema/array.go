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

	// Retrieve values validator JSON schema.
	var valuesSchema map[string]interface{}
	if v.Values.Validator != nil {
		b, err := ValidatorBuilder(v.Values.Validator)
		if err != nil {
			return nil, err
		}
		valuesSchema, err = b.BuildJSONSchema()
		if err != nil {
			return nil, err
		}
	} else {
		valuesSchema = map[string]interface{}{}
	}
	addFieldProperties(valuesSchema, v.Values)
	if len(valuesSchema) > 0 {
		m["items"] = valuesSchema
	}

	return m, nil
}
