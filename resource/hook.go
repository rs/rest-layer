package resource

import (
	"context"
	"errors"

	"github.com/rs/rest-layer/schema/query"
)

// FindEventHandler is an interface to be implemented by an event handler that
// want to be called before a find is performed on a resource. This interface is
// to be used with resource.Use() method.
type FindEventHandler interface {
	OnFind(ctx context.Context, q *query.Query) error
}

// FindEventHandlerFunc converts a function into a FindEventHandler.
type FindEventHandlerFunc func(ctx context.Context, q *query.Query) error

// OnFind implements FindEventHandler
func (e FindEventHandlerFunc) OnFind(ctx context.Context, q *query.Query) error {
	return e(ctx, q)
}

// FoundEventHandler is an interface to be implemented by an event handler that
// want to be called after a find has been performed on a resource. This
// interface is to be used with resource.Use() method.
type FoundEventHandler interface {
	OnFound(ctx context.Context, query *query.Query, list **ItemList, err *error)
}

// FoundEventHandlerFunc converts a function into a FoundEventHandler.
type FoundEventHandlerFunc func(ctx context.Context, q *query.Query, list **ItemList, err *error)

// OnFound implements FoundEventHandler
func (e FoundEventHandlerFunc) OnFound(ctx context.Context, q *query.Query, list **ItemList, err *error) {
	e(ctx, q, list, err)
}

// GetEventHandler is an interface to be implemented by an event handler that
// want to be called before a get is performed on a resource. This interface is
// to be used with resource.Use() method.
type GetEventHandler interface {
	OnGet(ctx context.Context, id interface{}) error
}

// GetEventHandlerFunc converts a function into a GetEventHandler.
type GetEventHandlerFunc func(ctx context.Context, id interface{}) error

// OnGet implements GetEventHandler
func (e GetEventHandlerFunc) OnGet(ctx context.Context, id interface{}) error {
	return e(ctx, id)
}

// GotEventHandler is an interface to be implemented by an event handler that
// want to be called after a get has been performed on a resource. This
// interface is to be used with resource.Use() method.
type GotEventHandler interface {
	OnGot(ctx context.Context, item **Item, err *error)
}

// GotEventHandlerFunc converts a function into a FoundEventHandler.
type GotEventHandlerFunc func(ctx context.Context, item **Item, err *error)

// OnGot implements GotEventHandler
func (e GotEventHandlerFunc) OnGot(ctx context.Context, item **Item, err *error) {
	e(ctx, item, err)
}

// InsertEventHandler is an interface to be implemented by an event handler that
// want to be called before an item is inserted on a resource. This interface is
// to be used with resource.Use() method.
type InsertEventHandler interface {
	OnInsert(ctx context.Context, items []*Item) error
}

// InsertEventHandlerFunc converts a function into a GetEventHandler.
type InsertEventHandlerFunc func(ctx context.Context, items []*Item) error

// OnInsert implements InsertEventHandler
func (e InsertEventHandlerFunc) OnInsert(ctx context.Context, items []*Item) error {
	return e(ctx, items)
}

// InsertedEventHandler is an interface to be implemented by an event handler
// that want to be called before an item has been inserted on a resource. This
// interface is to be used with resource.Use() method.
type InsertedEventHandler interface {
	OnInserted(ctx context.Context, items []*Item, err *error)
}

// InsertedEventHandlerFunc converts a function into a FoundEventHandler.
type InsertedEventHandlerFunc func(ctx context.Context, items []*Item, err *error)

// OnInserted implements InsertedEventHandler
func (e InsertedEventHandlerFunc) OnInserted(ctx context.Context, items []*Item, err *error) {
	e(ctx, items, err)
}

// UpdateEventHandler is an interface to be implemented by an event handler that
// want to be called before an item is updated for a resource. This interface is
// to be used with resource.Use() method.
type UpdateEventHandler interface {
	OnUpdate(ctx context.Context, item *Item, original *Item) error
}

// UpdateEventHandlerFunc converts a function into a GetEventHandler.
type UpdateEventHandlerFunc func(ctx context.Context, item *Item, original *Item) error

// OnUpdate implements UpdateEventHandler
func (e UpdateEventHandlerFunc) OnUpdate(ctx context.Context, item *Item, original *Item) error {
	return e(ctx, item, original)
}

// UpdatedEventHandler is an interface to be implemented by an event handler
// that want to be called before an item has been updated for a resource. This
// interface is to be used with resource.Use() method.
type UpdatedEventHandler interface {
	OnUpdated(ctx context.Context, item *Item, original *Item, err *error)
}

// UpdatedEventHandlerFunc converts a function into a FoundEventHandler.
type UpdatedEventHandlerFunc func(ctx context.Context, item *Item, original *Item, err *error)

// OnUpdated implements UpdatedEventHandler
func (e UpdatedEventHandlerFunc) OnUpdated(ctx context.Context, item *Item, original *Item, err *error) {
	e(ctx, item, original, err)
}

// DeleteEventHandler is an interface to be implemented by an event handler that
// want to be called before an item is deleted on a resource. This interface is
// to be used with resource.Use() method.
type DeleteEventHandler interface {
	OnDelete(ctx context.Context, item *Item) error
}

// DeleteEventHandlerFunc converts a function into a GetEventHandler.
type DeleteEventHandlerFunc func(ctx context.Context, item *Item) error

// OnDelete implements DeleteEventHandler
func (e DeleteEventHandlerFunc) OnDelete(ctx context.Context, item *Item) error {
	return e(ctx, item)
}

// DeletedEventHandler is an interface to be implemented by an event handler that
// want to be called before an item has been deleted on a resource. This interface is
// to be used with resource.Use() method.
type DeletedEventHandler interface {
	OnDeleted(ctx context.Context, item *Item, err *error)
}

// DeletedEventHandlerFunc converts a function into a FoundEventHandler.
type DeletedEventHandlerFunc func(ctx context.Context, item *Item, err *error)

// OnDeleted implements DeletedEventHandler
func (e DeletedEventHandlerFunc) OnDeleted(ctx context.Context, item *Item, err *error) {
	e(ctx, item, err)
}

// ClearEventHandler is an interface to be implemented by an event handler that
// want to be called before a resource is cleared. This interface is to be used
// with resource.Use() method.
type ClearEventHandler interface {
	OnClear(ctx context.Context, q *query.Query) error
}

// ClearEventHandlerFunc converts a function into a GetEventHandler.
type ClearEventHandlerFunc func(ctx context.Context, q *query.Query) error

// OnClear implements ClearEventHandler
func (e ClearEventHandlerFunc) OnClear(ctx context.Context, q *query.Query) error {
	return e(ctx, q)
}

// ClearedEventHandler is an interface to be implemented by an event handler
// that want to be called after a resource has been cleared. This interface is
// to be used with resource.Use() method.
type ClearedEventHandler interface {
	OnCleared(ctx context.Context, q *query.Query, deleted *int, err *error)
}

// ClearedEventHandlerFunc converts a function into a FoundEventHandler.
type ClearedEventHandlerFunc func(ctx context.Context, q *query.Query, deleted *int, err *error)

// OnCleared implements ClearedEventHandler
func (e ClearedEventHandlerFunc) OnCleared(ctx context.Context, q *query.Query, deleted *int, err *error) {
	e(ctx, q, deleted, err)
}

type eventHandler struct {
	onFindH     []FindEventHandler
	onFoundH    []FoundEventHandler
	onGetH      []GetEventHandler
	onGotH      []GotEventHandler
	onInsertH   []InsertEventHandler
	onInsertedH []InsertedEventHandler
	onUpdateH   []UpdateEventHandler
	onUpdatedH  []UpdatedEventHandler
	onDeleteH   []DeleteEventHandler
	onDeletedH  []DeletedEventHandler
	onClearH    []ClearEventHandler
	onClearedH  []ClearedEventHandler
}

func (h *eventHandler) use(e interface{}) error {
	found := false
	if e, ok := e.(FindEventHandler); ok {
		h.onFindH = append(h.onFindH, e)
		found = true
	}
	if e, ok := e.(FoundEventHandler); ok {
		h.onFoundH = append(h.onFoundH, e)
		found = true
	}
	if e, ok := e.(GetEventHandler); ok {
		h.onGetH = append(h.onGetH, e)
		found = true
	}
	if e, ok := e.(GotEventHandler); ok {
		h.onGotH = append(h.onGotH, e)
		found = true
	}
	if e, ok := e.(InsertEventHandler); ok {
		h.onInsertH = append(h.onInsertH, e)
		found = true
	}
	if e, ok := e.(InsertedEventHandler); ok {
		h.onInsertedH = append(h.onInsertedH, e)
		found = true
	}
	if e, ok := e.(UpdateEventHandler); ok {
		h.onUpdateH = append(h.onUpdateH, e)
		found = true
	}
	if e, ok := e.(UpdatedEventHandler); ok {
		h.onUpdatedH = append(h.onUpdatedH, e)
		found = true
	}
	if e, ok := e.(DeleteEventHandler); ok {
		h.onDeleteH = append(h.onDeleteH, e)
		found = true
	}
	if e, ok := e.(DeletedEventHandler); ok {
		h.onDeletedH = append(h.onDeletedH, e)
		found = true
	}
	if e, ok := e.(ClearEventHandler); ok {
		h.onClearH = append(h.onClearH, e)
		found = true
	}
	if e, ok := e.(ClearedEventHandler); ok {
		h.onClearedH = append(h.onClearedH, e)
		found = true
	}
	if !found {
		return errors.New("does not implement any event handler interface")
	}
	return nil
}

func (h *eventHandler) onFind(ctx context.Context, q *query.Query) error {
	for _, e := range h.onFindH {
		if err := e.OnFind(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (h *eventHandler) onFound(ctx context.Context, q *query.Query, list **ItemList, err *error) {
	for _, e := range h.onFoundH {
		e.OnFound(ctx, q, list, err)
	}
}

func (h *eventHandler) onGet(ctx context.Context, id interface{}) error {
	for _, e := range h.onGetH {
		if err := e.OnGet(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (h *eventHandler) onGot(ctx context.Context, item **Item, err *error) {
	for _, e := range h.onGotH {
		e.OnGot(ctx, item, err)
	}
}

func (h *eventHandler) onInsert(ctx context.Context, items []*Item) error {
	for _, e := range h.onInsertH {
		if err := e.OnInsert(ctx, items); err != nil {
			return err
		}
	}
	return nil
}

func (h *eventHandler) onInserted(ctx context.Context, items []*Item, err *error) {
	for _, e := range h.onInsertedH {
		e.OnInserted(ctx, items, err)
	}
}

func (h *eventHandler) onUpdate(ctx context.Context, item *Item, original *Item) error {
	for _, e := range h.onUpdateH {
		if err := e.OnUpdate(ctx, item, original); err != nil {
			return err
		}
	}
	return nil
}

func (h *eventHandler) onUpdated(ctx context.Context, item *Item, original *Item, err *error) {
	for _, e := range h.onUpdatedH {
		e.OnUpdated(ctx, item, original, err)
	}
}

func (h *eventHandler) onDelete(ctx context.Context, item *Item) error {
	for _, e := range h.onDeleteH {
		if err := e.OnDelete(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (h *eventHandler) onDeleted(ctx context.Context, item *Item, err *error) {
	for _, e := range h.onDeletedH {
		e.OnDeleted(ctx, item, err)
	}
}

func (h *eventHandler) onClear(ctx context.Context, q *query.Query) error {
	for _, e := range h.onClearH {
		if err := e.OnClear(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (h *eventHandler) onCleared(ctx context.Context, q *query.Query, deleted *int, err *error) {
	for _, e := range h.onClearedH {
		e.OnCleared(ctx, q, deleted, err)
	}
}
