package jsonschema

import (
	"math"

	"github.com/rs/rest-layer/schema"
)

type floatBuilder schema.Float

func (v floatBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	m := map[string]interface{}{
		"type": "number",
	}

	if len(v.Allowed) > 0 {
		m["enum"] = v.Allowed
	}
	if v.Boundaries != nil {
		addBoundariesProperties(m, v.Boundaries)
	}
	return m, nil
}

func addBoundariesProperties(m map[string]interface{}, b *schema.Boundaries) {
	if !math.IsNaN(b.Min) && !math.IsInf(b.Min, -1) {
		m["minimum"] = b.Min
	}
	if !math.IsNaN(b.Max) && !math.IsInf(b.Max, 1) {
		m["maximum"] = b.Max
	}
}
