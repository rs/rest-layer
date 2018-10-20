/*
Package query provides tools to query a schema defined by github.com/rs/schema.

A query is composed of the following elements:

  * Projection to define the format of the response.
  * Predicate to define selection criteria must match to be part of the result
    set.
  * Sort to define the order of the items.
  * Window to limit slice the result set.

The query package provides DLS to describe those elements as strings:

  * Projections uses a subset of GraphQL syntax. See ParseProjection for more
    info.
  * Predicate uses a subset of MongoDB query syntax. See ParsePredicate for more
    info.
  * Sort is a simple list of field separated by comas. See ParseSort for more
    info.

This package is part of the rest-layer project. See http://rest-layer.io for
full REST Layer documentation.
*/
package query

import "github.com/rs/rest-layer/schema"

// Query defines the criteria of a query to be applied on a resource validated
// by a schema.Schema.
type Query struct {
	// Projection is the list of fields from the items of the result that should
	// be included in the query response. A projected field can be aliased or
	// given parameters to be passed to per field transformation filters. A
	// projection is hierarchical allow projection of deep structures.
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
	if err := q.Predicate.Prepare(validator); err != nil {
		return err
	}
	if err := q.Sort.Validate(validator); err != nil {
		return err
	}
	return nil
}
