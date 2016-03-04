package schema

import "github.com/rs/xid"

var (
	// NewID is a field hook handler that generates a new unique id if none exist,
	// to be used in schema with OnInit.
	//
	// The generated ID is a Mongo like base64 object id (mgo/bson code has been embedded
	// into this function to prevent dep)
	NewID = func(value interface{}, _ []interface{}) interface{} {
		if value == nil {
			value = newID()
		}
		return value
	}

	// IDField is a common schema field configuration that generate an UUID for new item id.
	IDField = Field{
		Required:   true,
		ReadOnly:   true,
		OnInit:     &NewID,
		Filterable: true,
		Sortable:   true,
		Validator: &String{
			Regexp: "^[0-9a-v]{20}$",
		},
	}
)

// newID returns a new globally unique id using a copy of the mgo/bson algorithm.
func newID() string {
	return xid.New().String()
}
