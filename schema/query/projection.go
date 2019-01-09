package query

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/rest-layer/schema"
)

// Projection defines the list of fields that should be included into the
// returned payload, and how they should be represented. An empty Projection
// means all fields with no transformation.
type Projection []ProjectionField

// ProjectionField describes how a field should be represented in the returned
// payload.
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
func (p Projection) Validate(fg schema.FieldGetter) error {
	for _, pf := range p {
		if err := pf.Validate(fg); err != nil {
			return err
		}
	}
	return nil
}

// String output the projection in its DSL form.
func (p Projection) String() string {
	ps := make([]string, 0, len(p))
	for _, pf := range p {
		ps = append(ps, pf.String())
	}
	return strings.Join(ps, ",")
}

// String output the projection field in its DSL form.
func (pf ProjectionField) String() string {
	buf := &bytes.Buffer{}
	if pf.Alias != "" {
		buf.WriteString(pf.Alias)
		buf.WriteByte(':')
	}
	buf.WriteString(pf.Name)
	if len(pf.Params) > 0 {
		buf.WriteByte('(')
		names := make([]string, 0, len(pf.Params))
		for name := range pf.Params {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			value := pf.Params[name]
			buf.WriteString(name)
			buf.WriteByte(':')
			switch v := value.(type) {
			case string:
				buf.WriteString(strconv.Quote(v))
			case float64:
				buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
			case bool:
				buf.WriteString(fmt.Sprintf("%t", v))
			default:
				buf.WriteString(fmt.Sprintf("%q", v))
			}
			buf.WriteByte(',')
		}
		buf.Truncate(buf.Len() - 1) // remove the trailing coma.
		buf.WriteByte(')')
	}
	if len(pf.Children) > 0 {
		buf.WriteByte('{')
		buf.WriteString(pf.Children.String())
		buf.WriteByte('}')
	}
	return buf.String()
}
