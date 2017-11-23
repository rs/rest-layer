package resource

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/rest-layer/schema/query"
	"github.com/stretchr/testify/assert"
)

func TestHookUseFind(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(FindEventHandlerFunc(func(ctx context.Context, q *query.Query) error {
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
	err = h.onFind(context.Background(), nil)
	assert.True(t, called)

	err = h.use(FindEventHandlerFunc(func(ctx context.Context, q *query.Query) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onFind(context.Background(), nil)
	assert.EqualError(t, err, "error")
}

func TestHookUseFound(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(FoundEventHandlerFunc(func(ctx context.Context, q *query.Query, list **ItemList, err *error) {
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
	h.onFound(context.Background(), nil, nil, nil)
	assert.True(t, called)

	err = h.use(FoundEventHandlerFunc(func(ctx context.Context, q *query.Query, list **ItemList, err *error) {
		*list = &ItemList{}
		*err = errors.New("error")
	}))
	assert.NoError(t, err)
	var list *ItemList
	err = nil
	h.onFound(context.Background(), nil, &list, &err)
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
	h.onGet(context.Background(), nil)
	assert.True(t, called)

	err = h.use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onGet(context.Background(), nil)
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
	h.onGot(context.Background(), nil, nil)
	assert.True(t, called)

	err = h.use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		*item = &Item{}
		*err = errors.New("error")
	}))
	assert.NoError(t, err)
	var item *Item
	err = nil
	h.onGot(context.Background(), &item, &err)
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
	h.onInsert(context.Background(), nil)
	assert.True(t, called)

	err = h.use(InsertEventHandlerFunc(func(ctx context.Context, items []*Item) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onInsert(context.Background(), nil)
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
	h.onInserted(context.Background(), nil, nil)
	assert.True(t, called)

	err = h.use(InsertedEventHandlerFunc(func(ctx context.Context, items []*Item, err *error) {
		*err = errors.New("error")
	}))
	assert.NoError(t, err)
	err = nil
	h.onInserted(context.Background(), nil, &err)
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
	h.onUpdate(context.Background(), nil, nil)
	assert.True(t, called)

	err = h.use(UpdateEventHandlerFunc(func(ctx context.Context, item *Item, original *Item) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onUpdate(context.Background(), nil, nil)
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
	h.onUpdated(context.Background(), nil, nil, nil)
	assert.True(t, called)

	err = h.use(UpdatedEventHandlerFunc(func(ctx context.Context, item *Item, original *Item, err *error) {
		*err = errors.New("error")
	}))
	assert.NoError(t, err)
	err = nil
	h.onUpdated(context.Background(), nil, nil, &err)
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
	h.onDelete(context.Background(), nil)
	assert.True(t, called)

	err = h.use(DeleteEventHandlerFunc(func(ctx context.Context, item *Item) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onDelete(context.Background(), nil)
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
	h.onDeleted(context.Background(), nil, nil)
	assert.True(t, called)

	err = h.use(DeletedEventHandlerFunc(func(ctx context.Context, item *Item, err *error) {
		*err = errors.New("error")
	}))
	assert.NoError(t, err)
	err = nil
	h.onDeleted(context.Background(), nil, &err)
	assert.EqualError(t, err, "error")
}

func TestHookUseClear(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(ClearEventHandlerFunc(func(ctx context.Context, q *query.Query) error {
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
	h.onClear(context.Background(), nil)
	assert.True(t, called)

	err = h.use(ClearEventHandlerFunc(func(ctx context.Context, q *query.Query) error {
		return errors.New("error")
	}))
	assert.NoError(t, err)
	err = h.onClear(context.Background(), nil)
	assert.EqualError(t, err, "error")
}

func TestHookUseCleared(t *testing.T) {
	h := eventHandler{}
	called := false
	err := h.use(ClearedEventHandlerFunc(func(ctx context.Context, q *query.Query, deleted *int, err *error) {
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
	h.onCleared(context.Background(), nil, &deleted, nil)
	assert.True(t, called)

	err = h.use(ClearedEventHandlerFunc(func(ctx context.Context, q *query.Query, deleted *int, err *error) {
		*err = errors.New("error")
		*deleted = 2
	}))
	assert.NoError(t, err)
	err = nil
	h.onCleared(context.Background(), nil, &deleted, &err)
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

type allHooks struct {
	numCalled *int
}

func (h allHooks) OnFind(ctx context.Context, q *query.Query) error {
	*h.numCalled++
	return nil
}

func (h allHooks) OnFound(ctx context.Context, query *query.Query, list **ItemList, err *error) {
	*h.numCalled++
}

func (h allHooks) OnGet(ctx context.Context, id interface{}) error {
	*h.numCalled++
	return nil
}

func (h allHooks) OnGot(ctx context.Context, item **Item, err *error) {
	*h.numCalled++
}

func (h allHooks) OnInsert(ctx context.Context, items []*Item) error {
	*h.numCalled++
	return nil
}

func (h allHooks) OnInserted(ctx context.Context, items []*Item, err *error) {
	*h.numCalled++
}

func (h allHooks) OnUpdate(ctx context.Context, item *Item, original *Item) error {
	*h.numCalled++
	return nil
}

func (h allHooks) OnUpdated(ctx context.Context, item *Item, original *Item, err *error) {
	*h.numCalled++
}

func (h allHooks) OnDelete(ctx context.Context, item *Item) error {
	*h.numCalled++
	return nil
}

func (h allHooks) OnDeleted(ctx context.Context, item *Item, err *error) {
	*h.numCalled++
}

func (h allHooks) OnClear(ctx context.Context, q *query.Query) error {
	*h.numCalled++
	return nil
}

func (h allHooks) OnCleared(ctx context.Context, q *query.Query, deleted *int, err *error) {
	*h.numCalled++
}

func TestHookUseAll(t *testing.T) {
	var numCalled int

	h := eventHandler{}
	if err := h.use(allHooks{&numCalled}); err != nil {
		t.Fatal("test setup failed, h.use(hook{}) returned unexpected error: ", err)
	}
	if len(h.onFoundH) != 1 {
		t.Error("len(h.onFoundH) != 1")
	}
	if len(h.onGetH) != 1 {
		t.Error("len(h.onGetH) != 1")
	}
	if len(h.onGotH) != 1 {
		t.Error("len(h.onGotH) != 1")
	}
	if len(h.onInsertH) != 1 {
		t.Error("len(h.onInsertH) != 1")
	}
	if len(h.onInsertedH) != 1 {
		t.Error("len(h.onInsertedH) != 1")
	}
	if len(h.onUpdateH) != 1 {
		t.Error("len(h.onUpdateH) != 1")
	}
	if len(h.onUpdatedH) != 1 {
		t.Error("len(h.onUpdatedH) != 1")
	}
	if len(h.onDeleteH) != 1 {
		t.Error("len(h.onDeleteH) != 1")
	}
	if len(h.onDeletedH) != 1 {
		t.Error("len(h.onDeletedH) != 1")
	}
	if len(h.onClearH) != 1 {
		t.Error("len(h.onClearH) != 1")
	}
	if len(h.onClearedH) != 1 {
		t.Error("len(h.onClearedH) != 1")
	}

	ctx := context.Background()
	ctxd := WithDisableHooks(ctx)

	t.Run("onFind(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onFind(ctx, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onFind(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onFind(ctxd, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})
	t.Run("onFound(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		h.onFound(ctx, nil, nil, nil)

		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onFound(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		h.onFound(ctxd, nil, nil, nil)

		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})

	t.Run("onGet(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onGet(ctx, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onGet(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onGet(ctxd, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})
	t.Run("onGot(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		h.onGot(ctx, nil, nil)

		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onGot(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		h.onGot(ctxd, nil, nil)

		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})

	t.Run("onInsert(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onInsert(ctx, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onInsert(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onInsert(ctxd, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})
	t.Run("onInserted(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		h.onInserted(ctx, nil, nil)

		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onInserted(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		h.onInserted(ctxd, nil, nil)

		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})

	t.Run("onUpdate(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onUpdate(ctx, nil, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onUpdate(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onUpdate(ctxd, nil, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})
	t.Run("onUpdated(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		h.onUpdated(ctx, nil, nil, nil)

		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onUpdated(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		h.onUpdated(ctxd, nil, nil, nil)

		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})

	t.Run("onDelete(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onDelete(ctx, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onDelete(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onDelete(ctxd, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})
	t.Run("onDeleted(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		h.onDeleted(ctx, nil, nil)

		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onDeleted(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		h.onDeleted(ctxd, nil, nil)

		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})

	t.Run("onClear(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onClear(ctx, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onClear(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		if err := h.onClear(ctxd, nil); err != nil {
			t.Error("calling hook resulted in error: ", err)
		}
		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})
	t.Run("onCleared(ctx,nil)", func(t *testing.T) {
		numCalled = 0
		h.onCleared(ctx, nil, nil, nil)

		if numCalled != 1 {
			t.Errorf("expected numCalled == 1, got %d", numCalled)
		}
	})
	t.Run("onCleared(WithDisableHooks(ctx),nil)", func(t *testing.T) {
		numCalled = 0
		h.onCleared(ctxd, nil, nil, nil)

		if numCalled != 0 {
			t.Errorf("expected numCalled == 0, got %d", numCalled)
		}
	})
}
