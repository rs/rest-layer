package jsonschema

import "github.com/rs/rest-layer/schema"

type stringBuilder schema.String

func (v stringBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	m := map[string]interface{}{
		"type": "string",
	}

	if v.Regexp != "" {
		m["pattern"] = v.Regexp
	}
	if len(v.Allowed) > 0 {
		m["enum"] = v.Allowed
	}
	if v.MinLen > 0 {
		m["minLength"] = v.MinLen
	}
	if v.MaxLen > 0 {
		m["maxLength"] = v.MaxLen
	}
	return m, nil
}
