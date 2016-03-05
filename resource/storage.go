package resource

import (
	"github.com/rs/rest-layer/schema"
	"golang.org/x/net/context"
)

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

type storageHandler interface {
	Storer
	MultiGetter
	Get(ctx context.Context, id interface{}) (item *Item, err error)
}

type storageWrapper struct {
	Storer
}

// Get get one item by its id. If item is not found, ErrNotFound error is returned
func (s storageWrapper) Get(ctx context.Context, id interface{}) (item *Item, err error) {
	items, err := s.MultiGet(ctx, []interface{}{id})
	if err == nil {
		if len(items) == 1 && items[0].ID == id {
			item = items[0]
		} else {
			err = ErrNotFound
		}
	}
	return
}

// MultiGet get some items by their id and return them in the same order. If one or more item(s)
// is not found, their slot in the response is set to nil.
func (s storageWrapper) MultiGet(ctx context.Context, ids []interface{}) (items []*Item, err error) {
	if s.Storer == nil {
		return nil, ErrNoStorage
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	var tmp []*Item
	if mg, ok := s.Storer.(MultiGetter); ok {
		// If native support, use it
		tmp, err = mg.MultiGet(ctx, ids)
	} else {
		// Otherwise, emulate MultiGetter with a Find query
		l := NewLookup()
		if len(ids) == 1 {
			l.AddQuery(schema.Query{
				schema.Equal{Field: "id", Value: ids[0]},
			})
		} else {
			v := make([]schema.Value, len(ids))
			for i, id := range ids {
				v[i] = schema.Value(id)
			}
			l.AddQuery(schema.Query{
				schema.In{Field: "id", Values: v},
			})
		}
		var list *ItemList
		list, err = s.Storer.Find(ctx, l, 1, len(ids))
		if list != nil {
			tmp = list.Items
		}
	}
	if err != nil {
		return nil, err
	}
	// Sort items as requested
	items = make([]*Item, len(ids))
	for i, id := range ids {
		for _, item := range tmp {
			if item.ID == id {
				items[i] = item
			}
		}
	}
	return items, nil
}

// Find tries to use storer MultiGet with some pattern or Find otherwise
func (s storageWrapper) Find(ctx context.Context, lookup *Lookup, page, perPage int) (list *ItemList, err error) {
	if s.Storer == nil {
		return nil, ErrNoStorage
	}
	if mg, ok := s.Storer.(MultiGetter); ok {
		// If storage supports MultiGetter interface, detect some common find pattern that could be
		// converted to multi get
		if q := lookup.Filter(); len(q) == 1 && page == 1 && len(lookup.Sort()) == 0 {
			switch op := q[0].(type) {
			case schema.Equal:
				// When query pattern is a single document request by its id, use the multi get API
				if id, ok := op.Value.(string); ok && op.Field == "id" && (perPage == 1 || perPage < 0) {
					return wrapMgetList(mg.MultiGet(ctx, []interface{}{id}))
				}
			case schema.In:
				// When query pattern is a list of documents request by their ids, use the multi get API
				if op.Field == "id" && perPage < 0 || perPage == len(op.Values) {
					return wrapMgetList(mg.MultiGet(ctx, valuesToInterface(op.Values)))
				}
			}
		}
	}
	return s.Storer.Find(ctx, lookup, page, perPage)
}

// wrapMgetList wraps a MultiGet response into a resource.ItemList response
func wrapMgetList(items []*Item, err error) (*ItemList, error) {
	if err != nil {
		return nil, err
	}
	list := &ItemList{Page: 1, Total: len(items), Items: items}
	return list, nil
}

func (s storageWrapper) Insert(ctx context.Context, items []*Item) (err error) {
	if s.Storer == nil {
		return ErrNoStorage
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return s.Storer.Insert(ctx, items)
}

func (s storageWrapper) Update(ctx context.Context, item *Item, original *Item) (err error) {
	if s.Storer == nil {
		return ErrNoStorage
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return s.Storer.Update(ctx, item, original)
}

func (s storageWrapper) Delete(ctx context.Context, item *Item) (err error) {
	if s.Storer == nil {
		return ErrNoStorage
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return s.Storer.Delete(ctx, item)
}

func (s storageWrapper) Clear(ctx context.Context, lookup *Lookup) (deleted int, err error) {
	if s.Storer == nil {
		return 0, ErrNoStorage
	}
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	return s.Storer.Clear(ctx, lookup)
}
