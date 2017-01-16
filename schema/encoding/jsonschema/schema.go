package jsonschema

import (
	"errors"
	"sort"

	"github.com/rs/rest-layer/schema"
)

var (
	// ErrNotImplemented is returned when the JSON schema encoding logic for a schema.FieldValidator has not (yet)
	// been implemented.
	ErrNotImplemented = errors.New("not implemented")
)

func addSchemaProperties(m map[string]interface{}, s *schema.Schema) (err error) {
	if s == nil {
		return
	}
	m["type"] = "object"
	if s.Description != "" {
		m["description"] = s.Description
	}
	m["additionalProperties"] = false
	if s.MinLen > 0 {
		m["minProperties"] = s.MinLen
	}
	if s.MaxLen > 0 {
		m["maxProperties"] = s.MaxLen
	}
	if len(s.Fields) > 0 {
		err = addFields(m, s.Fields)
	}

	return err
}

func addFields(m map[string]interface{}, fields schema.Fields) error {
	props := make(map[string]interface{}, len(fields))
	required := []string{}

	for fieldName, field := range fields {
		if field.Required {
			required = append(required, fieldName)
		}
		builder, err := ValidatorBuilder(field.Validator)
		if err != nil {
			return err
		}
		fieldMap, err := builder.BuildJSONSchema()
		if err != nil {
			return err
		}
		addFieldProperties(fieldMap, field)
		props[fieldName] = fieldMap
	}
	m["properties"] = props

	if len(required) > 0 {
		sort.Strings(required)
		m["required"] = required
	}
	return nil
}

func addFieldProperties(m map[string]interface{}, field schema.Field) {
	if field.Description != "" {
		m["description"] = field.Description
	}
	if field.ReadOnly {
		m["readOnly"] = field.ReadOnly
	}
	if field.Default != nil {
		m["default"] = field.Default
	}
}

// ValidatorBuilder type-casts v to a valid Builder implementation or returns an error.
func ValidatorBuilder(v schema.FieldValidator) (Builder, error) {
	if v == nil {
		return builderFunc(nilBuilder), nil
	}
	switch t := v.(type) {
	case Builder:
		return t, nil
	case *schema.Bool:
		return (*boolBuilder)(t), nil
	case *schema.String:
		return (*stringBuilder)(t), nil
	case *schema.Time:
		return (*timeBuilder)(t), nil
	case *schema.Integer:
		return (*integerBuilder)(t), nil
	case *schema.Float:
		return (*floatBuilder)(t), nil
	case *schema.Array:
		return (*arrayBuilder)(t), nil
	case *schema.Object:
		return (*objectBuilder)(t), nil
	default:
		return nil, ErrNotImplemented
	}
}

func nilBuilder() (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
