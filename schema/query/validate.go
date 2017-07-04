package query

import (
	"fmt"

	"github.com/rs/rest-layer/schema"
)

func validateExpressions(exps []Expression, validator schema.Validator) error {
	for _, exp := range exps {
		if err := exp.Validate(validator); err != nil {
			return err
		}
	}
	return nil
}

func getValidatorField(field string, validator schema.Validator) (f *schema.Field, err error) {
	f = validator.GetField(field)
	if f == nil {
		return f, fmt.Errorf("%s: unknown query field", field)
	}
	if !f.Filterable {
		return f, fmt.Errorf("%s: field is not filterable", field)
	}
	return
}

func validateField(field string, validator schema.Validator) error {
	_, err := getValidatorField(field, validator)
	return err
}

func validateValues(field string, values []Value, validator schema.Validator) error {
	f, err := getValidatorField(field, validator)
	if err != nil {
		return err
	}
	if f.Validator != nil {
		for _, v := range values {
			if _, err := f.Validator.Validate(v); err != nil {
				return fmt.Errorf("%s: invalid query expression `%#v': %v", field, v, err)
			}
		}
	}
	return nil
}

func validateValue(field string, value Value, validator schema.Validator) error {
	f, err := getValidatorField(field, validator)
	if err != nil {
		return err
	}
	if f.Validator != nil {
		if f.Validator != nil {
			if _, err := f.Validator.Validate(value); err != nil {
				return fmt.Errorf("%s: invalid query expression: %s", field, err)
			}
		}
	}
	return nil
}

func validateNumericValue(field string, value Value, op string, validator schema.Validator) error {
	f, err := getValidatorField(field, validator)
	if err != nil {
		return err
	}
	if f.Validator != nil {
		switch f.Validator.(type) {
		case *schema.Integer, *schema.Float, schema.Integer, schema.Float:
			if _, err := f.Validator.Validate(value); err != nil {
				return fmt.Errorf("%s: invalid query expression: %v", field, err)
			}
		default:
			return fmt.Errorf("%s: cannot apply %s operation on a non numerical field", field, op)
		}
	}
	return nil
}
