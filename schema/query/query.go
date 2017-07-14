package query

import "github.com/rs/rest-layer/schema"

type Query struct {
	// Projection is the list of fields from the items of the result that should
	// be included in the query response. A projected field can be aliased or
	// given parameters to be passed to per field transformation filters. A
	// projection is hierachical allow projection of deep structures.
	//
	// A DSL can be used to build the projection structure.
	Projection Projection

	// Predicate defines the criteria an item must meet in order to be
	// considered for inclusion in the result set.
	//
	// A DLS can be used to build a predicate from a MongoDB like expressions.
	Predicate Predicate

	// Sort is a list of fields or sub-fields to use for sorting the result set.
	Sort Sort

	// Window defines result set windowing using an offset and a limit. When
	// nil, the full result-set should be returned.
	Window *Window
}

// New creates a query from a projection, predicate and sort queries using
// their respective DSL notations. An optional window can be provided.
//
// Example:
//
//   New("foo{bar},baz:b", `{baz: "bar"}`, "foo.bar,-baz", Page(1, 10, 0))
//
// Select items with the foo field equal to bar, including only the foo.bar and
// baz fields in the result set, with the baz field aliased to b. The result is
// then sorted by the foo.bar field ascending and baz descending. The result is
// windowed on page 1 with 10 items per page, skiping no result.
func New(projection, predicate, sort string, window *Window) (*Query, error) {
	proj, err := ParseProjection(projection)
	if err != nil {
		return nil, err
	}
	pred, err := ParsePredicate(predicate)
	if err != nil {
		return nil, err
	}
	s, err := ParseSort(sort)
	if err != nil {
		return nil, err
	}
	return &Query{
		Projection: proj,
		Predicate:  pred,
		Sort:       s,
		Window:     window,
	}, nil
}

// Validate validates the query against the provided validator.
func (q *Query) Validate(validator schema.Validator) error {
	if err := q.Projection.Validate(validator); err != nil {
		return err
	}
	if err := q.Predicate.Validate(validator); err != nil {
		return err
	}
	if err := q.Sort.Validate(validator); err != nil {
		return err
	}
	return nil
}
