package resource

import (
	"context"

	"github.com/rs/rest-layer/schema/query"
)

// Storer defines the interface of an handler able to store and retrieve resources
type Storer interface {
	// Find searches for items in the backend store matching the q argument. The
	// Window of the query must be respected. If no items are found, an empty
	// list should be returned with no error.
	//
	// If the total number of item can't be computed for free, ItemList.Total
	// must be set to -1. Your Storer may implement the Counter interface to let
	// the user explicitly request the total.
	//
	// The whole query must be treated. If a query predicate operation or sort
	// is not implemented by the storage handler, a resource.ErrNotImplemented
	// must be returned.
	//
	// A storer must ignore the Projection part of the query and always return
	// the document in its entirety. Documents matching a given predicate might
	// be reused (i.e.: cached) with a different projection.
	//
	// If the fetching of the data is not immediate, the method must listen for
	// cancellation on the passed ctx. If the operation is stopped due to
	// context cancellation, the function must return the result of the
	// ctx.Err() method.
	Find(ctx context.Context, q *query.Query) (*ItemList, error)
	// Insert stores new items in the backend store. If any of the items does
	// already exist, no item should be inserted and a resource.ErrConflict must
	// be returned. The insertion of the items must be performed atomically. If
	// more than one item is provided and the backend store doesn't support
	// atomical insertion of several items, a resource.ErrNotImplemented must be
	// returned.
	//
	// If the storage of the data is not immediate, the method must listen for
	// cancellation on the passed ctx. If the operation is stopped due to
	// context cancellation, the function must return the result of the
	// ctx.Err() method.
	Insert(ctx context.Context, items []*Item) error
	// Update replace an item in the backend store by a new version. The
	// ResourceHandler must ensure that the original item exists in the database
	// and has the same Etag field. This check should be performed atomically.
	// If the original item is not found, a resource.ErrNotFound must be
	// returned. If the etags don't match, a resource.ErrConflict must be
	// returned.
	//
	// The item payload must be stored together with the etag and the updated
	// field. The item.ID and the payload["id"] is guarantied to be identical,
	// so there's not need to store both.
	//
	// If the storage of the data is not immediate, the method must listen for
	// cancellation on the passed ctx. If the operation is stopped due to
	// context cancellation, the function must return the result of the
	// ctx.Err() method.
	Update(ctx context.Context, item *Item, original *Item) error
	// Delete deletes the provided item by its ID. The Etag of the item stored
	// in the backend store must match the Etag of the provided item or a
	// resource.ErrConflict must be returned. This check should be performed
	// atomically.
	//
	// If the provided item were not present in the backend store, a
	// resource.ErrNotFound must be returned.
	//
	// If the removal of the data is not immediate, the method must listen for
	// cancellation on the passed ctx. If the operation is stopped due to
	// context cancellation, the function must return the result of the
	// ctx.Err() method.
	Delete(ctx context.Context, item *Item) error
	// Clear removes all items matching the query. When possible, the number of
	// items removed is returned, otherwise -1 is return as the first value.
	//
	// The whole query must be treated. If a query predicate operation or sort
	// is not implemented by the storage handler, a resource.ErrNotImplemented
	// must be returned.
	//
	// If the removal of the data is not immediate, the method must listen for
	// cancellation on the passed ctx. If the operation is stopped due to
	// context cancellation, the function must return the result of the
	// ctx.Err() method.
	Clear(ctx context.Context, q *query.Query) (int, error)
}

// MultiGetter is an optional interface a Storer can implement when the storage
// engine is able to perform optimized multi gets. REST Layer will automatically
// use MultiGet over Find whenever it's possible when a storage handler
// implements this interface.
type MultiGetter interface {
	// MultiGet retrieves items by their ids and return them an a list. If one or more
	// item(s) cannot be found, the method must not return a resource.ErrNotFound but
	// must just omit the item in the result.
	//
	// The items in the result are expected to match the order of the requested ids.
	MultiGet(ctx context.Context, ids []interface{}) ([]*Item, error)
}

// Counter is an optional interface a Storer can implement to provide a way to
// explicitly count the total number of elements a given query would return.
// This method is called by REST Layer when the storage handler returned -1 as
// ItemList.Total and the user (or configuration) explicitly request the total.
type Counter interface {
	// Count returns the total number of item in the collection given the
	// provided query filter.
	Count(ctx context.Context, q *query.Query) (int, error)
}

type storageHandler interface {
	Storer
	MultiGetter
	Counter
	Get(ctx context.Context, id interface{}) (item *Item, err error)
}

type storageWrapper struct {
	Storer
}

// Get get one item by its id. If item is not found, ErrNotFound error is
// returned
func (s storageWrapper) Get(ctx context.Context, id interface{}) (item *Item, err error) {
	items, err := s.MultiGet(ctx, []interface{}{id})
	if err == nil {
		if len(items) == 1 && items[0] != nil && items[0].ID == id {
			item = items[0]
		} else {
			err = ErrNotFound
		}
	}
	return
}

// MultiGet get some items by their id and return them in the same order. If one
// or more item(s) is not found, their slot in the response is set to nil.
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
		q := &query.Query{}
		if len(ids) == 1 {
			q.Predicate = query.Predicate{
				&query.Equal{Field: "id", Value: ids[0]},
			}
		} else {
			v := make([]query.Value, len(ids))
			for i, id := range ids {
				v[i] = query.Value(id)
			}
			q.Predicate = query.Predicate{
				&query.In{Field: "id", Values: v},
			}
		}
		q.Window = &query.Window{Limit: len(ids)}
		var list *ItemList
		list, err = s.Storer.Find(ctx, q)
		if list != nil {
			tmp = list.Items
		}
	}
	if err != nil {
		return nil, err
	}
	// Sort items as requested.
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

// Find tries to use storer MultiGet with some pattern or Find otherwise.
func (s storageWrapper) Find(ctx context.Context, q *query.Query) (list *ItemList, err error) {
	if s.Storer == nil {
		return nil, ErrNoStorage
	}
	if mg, ok := s.Storer.(MultiGetter); ok {
		// If storage supports MultiGetter interface, detect some common find
		// pattern that could be converted to multi get.
		if len(q.Predicate) == 1 && (q.Window == nil || q.Window.Offset == 0) && len(q.Sort) == 0 {
			switch op := q.Predicate[0].(type) {
			case *query.Equal:
				// When query pattern is a single document request by its id,
				// use the multi get API.
				if id, ok := op.Value.(string); ok && op.Field == "id" && (q.Window == nil || q.Window.Limit == 1) {
					return wrapMgetList(mg.MultiGet(ctx, []interface{}{id}))
				}
			case *query.In:
				// When query pattern is a list of documents request by their
				// ids, use the multi get API.
				if op.Field == "id" && (q.Window == nil || q.Window.Limit == len(op.Values)) {
					return wrapMgetList(mg.MultiGet(ctx, op.Values))
				}
			}
		}
	}
	return s.Storer.Find(ctx, q)
}

// wrapMgetList wraps a MultiGet response into a resource.ItemList response.
func wrapMgetList(items []*Item, err error) (*ItemList, error) {
	if err != nil {
		return nil, err
	}
	list := &ItemList{Offset: 0, Total: len(items), Items: items}
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

func (s storageWrapper) Clear(ctx context.Context, q *query.Query) (deleted int, err error) {
	if s.Storer == nil {
		return 0, ErrNoStorage
	}
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	return s.Storer.Clear(ctx, q)
}

func (s storageWrapper) Count(ctx context.Context, q *query.Query) (total int, err error) {
	if s.Storer == nil {
		return -1, ErrNoStorage
	}
	if ctx.Err() != nil {
		return -1, ctx.Err()
	}
	if c, ok := s.Storer.(Counter); ok {
		return c.Count(ctx, q)
	}
	return -1, ErrNotImplemented
}
