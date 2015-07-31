package rest

// ResourceHandler defines the interface of an handler able to manage the life of a resource
type ResourceHandler interface {
	// Find searches for items in the backend store matching the lookup argument. The
	// pagination argument must be respected. If no items are found, an empty list
	// should be returned with no error.
	//
	// If the total number of item can't be easily computed, ItemList.Total should
	// be set to -1. The requested page should be set to ItemList.Page.
	Find(lookup *Lookup, page, perPage int) (*ItemList, *Error)
	// Store stores an item to the backend store. If the original item is provided, the
	// ResourceHandler must ensure that the item exists in the database and has the same
	// Etag field. This check should be performed atomically. If the original item is not
	// found, a rest.NotFoundError must be returned. If the etags don't match, a
	// rest.ConflictError must be returned.
	//
	// The item payload must be stored together with the etag and the updated field using.
	// The item.ID and the payload["id"] is garantied to be identical, so there's not need
	// to store both.
	Store(item *Item, original *Item) *Error
	// Delete deletes the provided item by its ID. The Etag of the item stored in the
	// backend store must match the Etag of the provided item or a rest.ConflictError
	// must be returned. This check should be performed atomically.
	//
	// If the provided item were not present in the backend store, a rest.NotFoundError
	// must be returned.
	Delete(item *Item) *Error
	// Clear removes all items maching the lookup. When possible, the number of items
	// removed is returned, otherwise -1 is return as the first value.
	Clear(lookup *Lookup) (int, *Error)
}
