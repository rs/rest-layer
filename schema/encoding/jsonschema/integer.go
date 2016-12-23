package jsonschema

import "github.com/rs/rest-layer/schema"

type integerBuilder schema.Integer

func (v integerBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	m := map[string]interface{}{
		"type": "integer",
	}

	if len(v.Allowed) > 0 {
		m["enum"] = v.Allowed
	}
	if v.Boundaries != nil {
		addBoundariesProperties(m, v.Boundaries)
	}
	return m, nil
}
