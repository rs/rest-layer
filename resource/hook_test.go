package resource

import (
	"errors"
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"
)

func TestHookUseFind(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(FindEventHandlerFunc(func(ctx context.Context, lookup *Lookup, page, perPage int) error {
		called = true
		return nil
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 1)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
	err = h.onFind(nil, nil, 0, 0)
	assert.True(t, called)

	err = h.use(FindEventHandlerFunc(func(ctx context.Context, lookup *Lookup, page, perPage int) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onFind(nil, nil, 0, 0)
	assert.EqualError(t, err, "error")
}

func TestHookUseFound(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(FoundEventHandlerFunc(func(ctx context.Context, lookup *Lookup, list **ItemList, err *error) {
		called = true
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 1)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
	h.onFound(nil, nil, nil, nil)
	assert.True(t, called)

	err = h.use(FoundEventHandlerFunc(func(ctx context.Context, lookup *Lookup, list **ItemList, err *error) {
		*list = &ItemList{}
		*err = errors.New("error")
	}))
	assert.NoError(t, err)
	var list *ItemList
	err = nil
	h.onFound(nil, nil, &list, &err)
	assert.EqualError(t, err, "error")
	assert.NotNil(t, list)
}

func TestHookUseGet(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		called = true
		return nil
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 1)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
	h.onGet(nil, nil)
	assert.True(t, called)

	err = h.use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onGet(nil, nil)
	assert.EqualError(t, err, "error")
}

func TestHookUseGot(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		called = true
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 1)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
	h.onGot(nil, nil, nil)
	assert.True(t, called)

	err = h.use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		*item = &Item{}
		*err = errors.New("error")
	}))
	assert.NoError(t, err)
	var item *Item
	err = nil
	h.onGot(nil, &item, &err)
	assert.EqualError(t, err, "error")
	assert.NotNil(t, item)
}

func TestHookUseInsert(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(InsertEventHandlerFunc(func(ctx context.Context, items []*Item) error {
		called = true
		return nil
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 1)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
	h.onInsert(nil, nil)
	assert.True(t, called)

	err = h.use(InsertEventHandlerFunc(func(ctx context.Context, items []*Item) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onInsert(nil, nil)
	assert.EqualError(t, err, "error")
}

func TestHookUseInserted(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(InsertedEventHandlerFunc(func(ctx context.Context, items []*Item, err *error) {
		called = true
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 1)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
	h.onInserted(nil, nil, nil)
	assert.True(t, called)

	err = h.use(InsertedEventHandlerFunc(func(ctx context.Context, items []*Item, err *error) {
		*err = errors.New("error")
	}))
	assert.NoError(t, err)
	err = nil
	h.onInserted(nil, nil, &err)
	assert.EqualError(t, err, "error")
}

func TestHookUseUpdate(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(UpdateEventHandlerFunc(func(ctx context.Context, item *Item, original *Item) error {
		called = true
		return nil
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 1)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
	h.onUpdate(nil, nil, nil)
	assert.True(t, called)

	err = h.use(UpdateEventHandlerFunc(func(ctx context.Context, item *Item, original *Item) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onUpdate(nil, nil, nil)
	assert.EqualError(t, err, "error")
}

func TestHookUseUpdated(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(UpdatedEventHandlerFunc(func(ctx context.Context, item *Item, original *Item, err *error) {
		called = true
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 1)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
	h.onUpdated(nil, nil, nil, nil)
	assert.True(t, called)

	err = h.use(UpdatedEventHandlerFunc(func(ctx context.Context, item *Item, original *Item, err *error) {
		*err = errors.New("error")
	}))
	assert.NoError(t, err)
	err = nil
	h.onUpdated(nil, nil, nil, &err)
	assert.EqualError(t, err, "error")
}

func TestHookUseDelete(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(DeleteEventHandlerFunc(func(ctx context.Context, item *Item) error {
		called = true
		return nil
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 1)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
	h.onDelete(nil, nil)
	assert.True(t, called)

	err = h.use(DeleteEventHandlerFunc(func(ctx context.Context, item *Item) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onDelete(nil, nil)
	assert.EqualError(t, err, "error")
}

func TestHookUseDeleted(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(DeletedEventHandlerFunc(func(ctx context.Context, item *Item, err *error) {
		called = true
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 1)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
	h.onDeleted(nil, nil, nil)
	assert.True(t, called)

	err = h.use(DeletedEventHandlerFunc(func(ctx context.Context, item *Item, err *error) {
		*err = errors.New("error")
	}))
	assert.NoError(t, err)
	err = nil
	h.onDeleted(nil, nil, &err)
	assert.EqualError(t, err, "error")
}

func TestHookUseClear(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(ClearEventHandlerFunc(func(ctx context.Context, lookup *Lookup) error {
		called = true
		return nil
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 1)
	assert.Len(t, h.onClearedH, 0)
	h.onClear(nil, nil)
	assert.True(t, called)

	err = h.use(ClearEventHandlerFunc(func(ctx context.Context, lookup *Lookup) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onClear(nil, nil)
	assert.EqualError(t, err, "error")
}

func TestHookUseCleared(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(ClearedEventHandlerFunc(func(ctx context.Context, lookup *Lookup, deleted *int, err *error) {
		called = true
	}))
	assert.NoError(t, err)
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 1)
	deleted := 0
	h.onCleared(nil, nil, &deleted, nil)
	assert.True(t, called)

	err = h.use(ClearedEventHandlerFunc(func(ctx context.Context, lookup *Lookup, deleted *int, err *error) {
		*err = errors.New("error")
		*deleted = 2
	}))
	assert.NoError(t, err)
	err = nil
	h.onCleared(nil, nil, &deleted, &err)
	assert.EqualError(t, err, "error")
	assert.Equal(t, 2, deleted)
}

func TestHookUseNonEventHandler(t *testing.T) {
	h := eventHandler{}
	err := h.use("something else")
	assert.EqualError(t, err, "does not implement any event handler interface")
	assert.Len(t, h.onFindH, 0)
	assert.Len(t, h.onFoundH, 0)
	assert.Len(t, h.onGetH, 0)
	assert.Len(t, h.onGotH, 0)
	assert.Len(t, h.onInsertH, 0)
	assert.Len(t, h.onInsertedH, 0)
	assert.Len(t, h.onUpdateH, 0)
	assert.Len(t, h.onUpdatedH, 0)
	assert.Len(t, h.onDeleteH, 0)
	assert.Len(t, h.onDeletedH, 0)
	assert.Len(t, h.onClearH, 0)
	assert.Len(t, h.onClearedH, 0)
}
