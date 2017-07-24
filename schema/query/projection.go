package query

import (
	"github.com/rs/rest-layer/schema"
)

type Projection []ProjectionField

type ProjectionField struct {
	// Name is the name of the field as define in the resource's schema.
	Name string

	// Alias is the wanted name in the representation.
	Alias string

	// Params defines a list of params to be sent to the field's param handler
	// if any.
	Params map[string]interface{}

	// Children holds references to child projections if any.
	Children Projection
}

// Validate validates the projection against the provided validator.
func (p Projection) Validate(validator schema.Validator) error {
	for _, pf := range p {
		if err := pf.Validate(validator); err != nil {
			return err
		}
	}
	return nil
}

// String output the projection in its DSL form.
func (Projection) String() string {
	return "" // XXX
}
