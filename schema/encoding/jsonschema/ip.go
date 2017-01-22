package jsonschema

import "github.com/rs/rest-layer/schema"

type ipBuilder schema.IP

func (v ipBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	m := map[string]interface{}{
		"type": "string",
		"oneOf": []map[string]interface{}{
			{"format": "ipv4"},
			{"format": "ipv6"},
		},
	}
	return m, nil
}
