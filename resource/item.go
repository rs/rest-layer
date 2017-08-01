package resource

import (
	"errors"
	"strings"
	"time"
)

// Item represents an instance of an item.
type Item struct {
	// ID is used to uniquely identify the item in the resource collection.
	ID interface{}
	// ETag is an opaque identifier assigned by REST Layer to a specific version
	// of the item.
	//
	// This ETag is used perform conditional requests and to ensure storage
	// handler doesn't update an outdated version of the resource.
	ETag string
	// Updated stores the last time the item was updated. This field is used to
	// populate the Last-Modified header and to handle conditional requests.
	Updated time.Time
	// Payload the actual data of the item
	Payload map[string]interface{}
}

// ItemList represents a list of items
type ItemList struct {
	// Total defines the total number of items in the collection matching the
	// current context. If the storage handler cannot compute this value, -1 is
	// set.
	Total int
	// Offset is the index of the first item of the list in the global
	// collection.
	Offset int
	// Limit is the max number of items requested.
	Limit int
	// Items is the list of items contained in the current page given the
	// current context.
	Items []*Item
}

// NewItem creates a new item from a payload.
func NewItem(payload map[string]interface{}) (*Item, error) {
	id, found := payload["id"]
	if !found {
		return nil, errors.New("Missing ID field")
	}
	etag, err := genEtag(payload)
	if err != nil {
		return nil, err
	}
	item := &Item{
		ID:      id,
		ETag:    etag,
		Updated: time.Now(),
		Payload: payload,
	}
	return item, nil
}

// GetField returns the item's payload field by its name.
//
// A field name may use the dot notation to reference a sub field. A GetField on
// field.subfield is equivalent to item.Payload["field"]["subfield"].
func (i *Item) GetField(name string) interface{} {
	return getField(i.Payload, name)
}

// getField gets the value of a given field by supporting sub-field path. A get
// on field.subfield is equivalent to payload["field"]["subfield].
func getField(payload map[string]interface{}, name string) interface{} {
	// Split the name to get the current level name on first element and the
	// rest of the path as second element if dot notation is used (i.e.:
	// field.subfield.subsubfield -> field, subfield.subsubfield).
	path := strings.SplitN(name, ".", 2)
	if value, found := payload[path[0]]; found {
		if len(path) == 2 {
			if subPayload, ok := value.(map[string]interface{}); ok {
				// Check next level
				return getField(subPayload, path[1])
			}
			// The requested depth does not exist
			return nil
		}
		// Full path has been found
		return value
	}
	return nil
}
