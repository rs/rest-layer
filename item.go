package rest

import (
	"errors"
	"strings"
	"time"
)

// Item represents an instance of an item
type Item struct {
	ID      interface{}
	Etag    string
	Updated time.Time
	Payload map[string]interface{}
}

// ItemList represents a list of items
type ItemList struct {
	Total int
	Page  int
	Items []*Item
}

// NewItem creates a new item from a payload
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
		Etag:    etag,
		Updated: time.Now(),
		Payload: payload,
	}
	return item, nil
}

// GetField returns the item's payload field by its name.
//
// A field name may use the dot notation to refrence a sub field.
// A GetField on field.subfield is equivalent to item.Payload["field"]["subfield].
func (i Item) GetField(name string) interface{} {
	return getField(i.Payload, name)
}

// getField gets the value of a given field by supporting sub-field path.
// A get on field.subfield is equivalent to payload["field"]["subfield].
func getField(payload map[string]interface{}, name string) interface{} {
	// Split the name to get the current level name on first element and
	// the rest of the path as second element if dot notation is used
	// (i.e.: field.subfield.subsubfield -> field, subfield.subsubfield)
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
