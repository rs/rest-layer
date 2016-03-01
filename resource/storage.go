package resource

import "golang.org/x/net/context"

// Storer defines the interface of an handler able to store and retreive resources
type Storer interface {
	// Find searches for items in the backend store matching the lookup argument. The
	// pagination argument must be respected. If no items are found, an empty list
	// should be returned with no error.
	//
	// If the total number of item can't be easily computed, ItemList.Total should
	// be set to -1. The requested page should be set to ItemList.Page.
	//
	// The whole lookup query must be treated. If a query operation is not implemented
	// by the storage handler, a resource.ErrNotImplemented must be returned.
	//
	// If the fetching of the data is not immediate, the method must listen for cancellation
	// on the passed ctx. If the operation is stopped due to context cancellation, the
	// function must return the result of the ctx.Err() method.
	//
	// If you need to log something, use xlog.FromContext(ctx) to get a logger. See
	// https://github.com/rs/xlog for more info.
	Find(ctx context.Context, lookup *Lookup, page, perPage int) (*ItemList, error)
	// Insert stores new items in the backend store. If any of the items does already exist,
	// no item should be inserted and a resource.ErrConflict must be returned. The insertion
	// of the items must be performed atomically. If more than one item is provided and the
	// backend store doesn't support atomical insertion of several items, a
	// resource.ErrNotImplemented must be returned.
	//
	// If the storage of the data is not immediate, the method must listen for cancellation
	// on the passed ctx. If the operation is stopped due to context cancellation, the
	// function must return the result of the ctx.Err() method.
	//
	// If you need to log something, use xlog.FromContext(ctx) to get a logger. See
	// https://github.com/rs/xlog for more info.
	Insert(ctx context.Context, items []*Item) error
	// Update replace an item in the backend store by a new version. The ResourceHandler must
	// ensure that the original item exists in the database and has the same Etag field.
	// This check should be performed atomically. If the original item is not
	// found, a resource.ErrNotFound must be returned. If the etags don't match, a
	// resource.ErrConflict must be returned.
	//
	// The item payload must be stored together with the etag and the updated field.
	// The item.ID and the payload["id"] is garantied to be identical, so there's not need
	// to store both.
	//
	// If the storage of the data is not immediate, the method must listen for cancellation
	// on the passed ctx. If the operation is stopped due to context cancellation, the
	// function must return the result of the ctx.Err() method.
	//
	// If you need to log something, use xlog.FromContext(ctx) to get a logger. See
	// https://github.com/rs/xlog for more info.
	Update(ctx context.Context, item *Item, original *Item) error
	// Delete deletes the provided item by its ID. The Etag of the item stored in the
	// backend store must match the Etag of the provided item or a resource.ErrConflict
	// must be returned. This check should be performed atomically.
	//
	// If the provided item were not present in the backend store, a resource.ErrNotFound
	// must be returned.
	//
	// If the removal of the data is not immediate, the method must listen for cancellation
	// on the passed ctx. If the operation is stopped due to context cancellation, the
	// function must return the result of the ctx.Err() method.
	//
	// If you need to log something, use xlog.FromContext(ctx) to get a logger. See
	// https://github.com/rs/xlog for more info.
	Delete(ctx context.Context, item *Item) error
	// Clear removes all items maching the lookup. When possible, the number of items
	// removed is returned, otherwise -1 is return as the first value.
	//
	// The whole lookup query must be treated. If a query operation is not implemented
	// by the storage handler, a resource.ErrNotImplemented must be returned.
	//
	// If the removal of the data is not immediate, the method must listen for cancellation
	// on the passed ctx. If the operation is stopped due to context cancellation, the
	// function must return the result of the ctx.Err() method.
	//
	// If you need to log something, use xlog.FromContext(ctx) to get a logger. See
	// https://github.com/rs/xlog for more info.
	Clear(ctx context.Context, lookup *Lookup) (int, error)
}

// MultiGetter is an optional interface a Storer can implement when the storage engine is
// able to perform optimized multi gets. REST Layer will automatically use MultiGet over Find
// whenever it's possible when a storage handler implements this interface.
type MultiGetter interface {
	// MultiGet retreives items by their ids and return them an a list. If one or more
	// item(s) cannot be found, the method must not return a resource.ErrNotFound but
	// must just omit the item in the result.
	//
	// The items in the result are expected to match the order of the requested ids.
	MultiGet(ctx context.Context, ids []interface{}) ([]*Item, error)
}
