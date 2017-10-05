package jsonschema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/rest-layer/schema"
)

type dictBuilder schema.Dict

var (
	//ErrKeysValidatorNotSupported is returned when Dict.KeysValidator is not a
	//*schema.String instance or nil.
	ErrKeysValidatorNotSupported = errors.New("KeysValidator type not supported")
)

func (v dictBuilder) BuildJSONSchema() (map[string]interface{}, error) {
	m := map[string]interface{}{
		"type": "object",
	}
	var patterns []string

	// Retrieve keys validator pattern(s).
	switch kv := v.KeysValidator.(type) {
	case *schema.String:
		if len(kv.Allowed) > 0 {
			patterns = append(patterns,
				fmt.Sprintf("^(%s)$", strings.Join(kv.Allowed, "|")),
			)

		}
		if kv.MaxLen > 0 || kv.MinLen > 0 {
			if kv.MaxLen == kv.MinLen {
				patterns = append(patterns,
					fmt.Sprintf("^.{%d}$", kv.MinLen),
				)
			} else if kv.MaxLen > 0 {
				patterns = append(patterns,
					fmt.Sprintf("^.{%d,%d}$", kv.MinLen, kv.MaxLen),
				)
			} else {
				patterns = append(patterns,
					fmt.Sprintf("^.{%d,}$", kv.MinLen),
				)
			}
		}
		if kv.Regexp != "" {
			patterns = append(patterns, kv.Regexp)
		}
	case nil:
	default:
		return nil, ErrKeysValidatorNotSupported
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

	// Compose JSON Schema.
	switch len(patterns) {
	case 0:
		if len(valuesSchema) > 0 {
			m["additionalProperties"] = valuesSchema
		} else {
			m["additionalProperties"] = true
		}
	case 1:
		m["additionalProperties"] = false
		m["patternProperties"] = map[string]interface{}{
			patterns[0]: valuesSchema,
		}
	default:
		// With the lack of logical AND in O(n) regex implementations (e.g. the
		// Go implementation), we have to build an allOff clause (with
		// duplicated schemas for values validation) whenever multiple key
		// patterns are found.
		allOf := make([]map[string]map[string]interface{}, 0, len(patterns))
		for i := range patterns {
			allOf = append(allOf, map[string]map[string]interface{}{
				"patternProperties": {
					patterns[i]: valuesSchema,
				},
			})
		}
		m["additionalProperties"] = false
		m["allOf"] = allOf
	}

	return m, nil
}
