package rest

import "golang.org/x/net/context"

// ResourceHandler defines the interface of an handler able to manage the life of a resource
type ResourceHandler interface {
	// Find searches for items in the backend store matching the lookup argument. The
	// pagination argument must be respected. If no items are found, an empty list
	// should be returned with no error.
	//
	// If the total number of item can't be easily computed, ItemList.Total should
	// be set to -1. The requested page should be set to ItemList.Page.
	//
	// If the fetching of the data is not immediate, the method must listen for cancellation
	// on the passed ctx. If the operation is stopped due to context cancellation, the
	// function must return the result of the rest.ContextError() with the ctx.Err() as
	// argument.
	Find(ctx context.Context, lookup Lookup, page, perPage int) (*ItemList, *Error)
	// Insert stores new items in the backend store. If any of the items does already exist,
	// no item should be inserted and a rest.ConflictError must be returned. The insertion
	// of the items must be performed atomically. If more than one item is provided and the
	// backend store doesn't support atomical insertion of several items, a
	// rest.NotImplementedError must be returned.
	//
	// If the fetching of the data is not immediate, the method must listen for cancellation
	// on the passed ctx. If the operation is stopped due to context cancellation, the
	// function must return the result of the rest.ContextError() with the ctx.Err() as
	// argument.
	Insert(ctx context.Context, items []*Item) *Error
	// Update replace an item in the backend store by a new version. The ResourceHandler must
	// ensure that the original item exists in the database and has the same Etag field.
	// This check should be performed atomically. If the original item is not
	// found, a rest.NotFoundError must be returned. If the etags don't match, a
	// rest.ConflictError must be returned.
	//
	// The item payload must be stored together with the etag and the updated field.
	// The item.ID and the payload["id"] is garantied to be identical, so there's not need
	// to store both.
	//
	// If the fetching of the data is not immediate, the method must listen for cancellation
	// on the passed ctx. If the operation is stopped due to context cancellation, the
	// function must return the result of the rest.ContextError() with the ctx.Err() as
	// argument.
	Update(ctx context.Context, item *Item, original *Item) *Error
	// Delete deletes the provided item by its ID. The Etag of the item stored in the
	// backend store must match the Etag of the provided item or a rest.ConflictError
	// must be returned. This check should be performed atomically.
	//
	// If the provided item were not present in the backend store, a rest.NotFoundError
	// must be returned.
	//
	// If the fetching of the data is not immediate, the method must listen for cancellation
	// on the passed ctx. If the operation is stopped due to context cancellation, the
	// function must return the result of the rest.ContextError() with the ctx.Err() as
	// argument.
	Delete(ctx context.Context, item *Item) *Error
	// Clear removes all items maching the lookup. When possible, the number of items
	// removed is returned, otherwise -1 is return as the first value.
	//
	// If the fetching of the data is not immediate, the method must listen for cancellation
	// on the passed ctx. If the operation is stopped due to context cancellation, the
	// function must return the result of the rest.ContextError() with the ctx.Err() as
	// argument.
	Clear(ctx context.Context, lookup Lookup) (int, *Error)
}
