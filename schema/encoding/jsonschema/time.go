package jsonschema

import "github.com/rs/rest-layer/schema"

type timeBuilder schema.Time

func (v timeBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":   "string",
		"format": "date-time",
	}, nil
}
