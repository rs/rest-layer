package schema

import (
	"context"

	"github.com/rs/xid"
)

var (
	// NewID is a field hook handler that generates a new globally unique id if
	// none exist, to be used in schema with OnInit.
	//
	// The generated ID is a Mongo like base64 object id (mgo/bson code has been
	// embedded into this function to prevent dep).
	NewID = func(ctx context.Context, value interface{}) interface{} {
		if value == nil {
			value = newID()
		}
		return value
	}

	// IDField is a common schema field configuration that generate an globally
	// unique id for new item id.
	IDField = Field{
		Description: "The item's id",
		Required:    true,
		ReadOnly:    true,
		OnInit:      NewID,
		Filterable:  true,
		Sortable:    true,
		Validator: &String{
			// This regexp matches a base32 id
			Regexp: "^[0-9a-v]{20}$",
		},
	}
)

// newID returns a new globally unique id using a copy of the mgo/bson
// algorithm.
func newID() string {
	return xid.New().String()
}
