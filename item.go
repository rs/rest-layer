package rest

import (
	"errors"
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
