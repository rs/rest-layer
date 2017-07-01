package jsonschema

import "github.com/rs/rest-layer/schema"

type urlBuilder schema.URL

func (v urlBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	// TODO: Currently the JSON Schema representation ignores any validation
	// configuration set in schema.URL.
	m := map[string]interface{}{
		"type":   "string",
		"format": "uri",
	}
	return m, nil
}
