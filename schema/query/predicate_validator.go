package query

import (
	"fmt"

	"github.com/rs/rest-layer/schema"
)

func prepareExpressions(exps []Expression, validator schema.Validator) error {
	for _, exp := range exps {
		if err := exp.Prepare(validator); err != nil {
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

func prepareValues(field string, values []Value, validator schema.Validator) error {
	f, err := getValidatorField(field, validator)
	if err != nil {
		return err
	}
	if f.Validator == nil {
		return nil
	}
	// use default validation method
	validateFunc := f.Validator.Validate
	qv, ok := f.Validator.(schema.FieldQueryValidator)
	if ok {
		// if there is explicit validator for quering
		validateFunc = qv.ValidateQuery
	}
	for i, v := range values {
		nv, err := validateFunc(v)
		if err != nil {
			return fmt.Errorf("%s: invalid query expression `%#v': %v", field, v, err)
		}
		values[i] = nv
	}
	return nil
}

func prepareValue(field string, value Value, validator schema.Validator) (Value, error) {
	f, err := getValidatorField(field, validator)
	if err != nil {
		return nil, err
	}
	if f.Validator == nil {
		return value, nil
	}
	// use default validation method
	validateFunc := f.Validator.Validate
	qv, ok := f.Validator.(schema.FieldQueryValidator)
	if ok {
		// if there is explicit validator for quering
		validateFunc = qv.ValidateQuery
	}
	nv, err := validateFunc(value)
	if err != nil {
		return nil, fmt.Errorf("%s: invalid query expression: %s", field, err)
	}
	return nv, nil
}

func getLessFunc(field string, validator schema.Validator) (schema.LessFunc, error) {
	f, err := getValidatorField(field, validator)
	if err != nil {
		return nil, err
	}
	fc, ok := f.Validator.(schema.FieldComparator)
	if !ok {
		return nil, fmt.Errorf("%s: not-comparable", field)
	}
	less := fc.LessFunc()
	if less == nil {
		return nil, fmt.Errorf("%s: not-comparable", field)
	}
	return less, nil
}
