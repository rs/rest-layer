package resource

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
)

type testStorer struct {
	find   func(ctx context.Context, lookup *Lookup, offset, limit int) (*ItemList, error)
	insert func(ctx context.Context, items []*Item) error
	update func(ctx context.Context, item *Item, original *Item) error
	delete func(ctx context.Context, item *Item) error
	clear  func(ctx context.Context, lookup *Lookup) (int, error)
}

type testMStorer struct {
	testStorer
	multiGet func(ctx context.Context, ids []interface{}) ([]*Item, error)
}

func (s testStorer) Find(ctx context.Context, lookup *Lookup, offset, limit int) (*ItemList, error) {
	return s.find(ctx, lookup, offset, limit)
}
func (s testStorer) Insert(ctx context.Context, items []*Item) error {
	return s.insert(ctx, items)
}
func (s testStorer) Update(ctx context.Context, item *Item, original *Item) error {
	return s.update(ctx, item, original)
}
func (s testStorer) Delete(ctx context.Context, item *Item) error {
	return s.delete(ctx, item)
}
func (s testStorer) Clear(ctx context.Context, lookup *Lookup) (int, error) {
	return s.clear(ctx, lookup)
}
func (s testMStorer) MultiGet(ctx context.Context, ids []interface{}) ([]*Item, error) {
	return s.multiGet(ctx, ids)
}

func newTestStorer() *testStorer {
	return &testStorer{
		find: func(ctx context.Context, lookup *Lookup, offset, limit int) (*ItemList, error) {
			return &ItemList{}, nil
		},
		insert: func(ctx context.Context, items []*Item) error {
			return nil
		},
		update: func(ctx context.Context, item *Item, original *Item) error {
			return nil
		},
		delete: func(ctx context.Context, item *Item) error {
			return nil
		},
		clear: func(ctx context.Context, lookup *Lookup) (int, error) {
			return 0, nil
		},
	}
}

func newTestMStorer() *testMStorer {
	return &testMStorer{
		testStorer: *newTestStorer(),
		multiGet: func(ctx context.Context, ids []interface{}) ([]*Item, error) {
			return []*Item{}, nil
		},
	}
}

/*
 * Get
 */

func TestResourceGet(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.multiGet = func(ctx context.Context, ids []interface{}) ([]*Item, error) {
		handler = true
		return []*Item{{ID: 1}}, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		preHook = true
		assert.Equal(t, 1, id)
		return nil
	}))
	r.Use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		postHook = true
		assert.Equal(t, &Item{ID: 1}, *item)
		assert.NoError(t, *err)
	}))
	ctx := context.Background()
	_, err := r.Get(ctx, 1)
	assert.NoError(t, err)
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceGetPostHookOverwrite(t *testing.T) {
	i := NewIndex()
	s := newTestMStorer()
	s.multiGet = func(ctx context.Context, ids []interface{}) ([]*Item, error) {
		return []*Item{{ID: 1}}, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		*item = &Item{ID: 2}
	}))
	ctx := context.Background()
	item, err := r.Get(ctx, 1)
	assert.Equal(t, &Item{ID: 2}, item)
	assert.NoError(t, err)
}

func TestResourceGetError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.multiGet = func(ctx context.Context, ids []interface{}) ([]*Item, error) {
		handler = true
		return nil, errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		preHook = true
		assert.Equal(t, 1, id)
		return nil
	}))
	r.Use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		postHook = true
		assert.Nil(t, *item)
		assert.EqualError(t, *err, "storer error")
	}))
	ctx := context.Background()
	_, err := r.Get(ctx, 1)
	assert.EqualError(t, err, "storer error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceGetPreHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.multiGet = func(ctx context.Context, ids []interface{}) ([]*Item, error) {
		handler = true
		return nil, errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		preHook = true
		return errors.New("pre hook error")
	}))
	r.Use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		postHook = true
		assert.Nil(t, *item)
		assert.EqualError(t, *err, "pre hook error")
	}))
	ctx := context.Background()
	_, err := r.Get(ctx, 1)
	assert.EqualError(t, err, "pre hook error")
	assert.True(t, preHook)
	assert.False(t, handler)
	assert.True(t, postHook)
}

func TestResourceGetPostHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.multiGet = func(ctx context.Context, ids []interface{}) ([]*Item, error) {
		handler = true
		return []*Item{{ID: 1}}, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		preHook = true
		assert.Equal(t, 1, id)
		return nil
	}))
	r.Use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		postHook = true
		assert.Equal(t, &Item{ID: 1}, *item)
		assert.NoError(t, *err)
		*err = errors.New("post hook error")
	}))
	ctx := context.Background()
	_, err := r.Get(ctx, 1)
	assert.EqualError(t, err, "post hook error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

/*
 * MultiGet
 */

func TestResourceMultiGet(t *testing.T) {
	var preHook, postHook, handler int
	i := NewIndex()
	s := newTestMStorer()
	s.multiGet = func(ctx context.Context, ids []interface{}) ([]*Item, error) {
		handler++
		return []*Item{{ID: 1}}, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		preHook++
		assert.Equal(t, 1, id)
		return nil
	}))
	r.Use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		postHook++
		assert.Equal(t, &Item{ID: 1}, *item)
		assert.NoError(t, *err)
	}))
	ctx := context.Background()
	items, err := r.MultiGet(ctx, []interface{}{1, 1})
	assert.Len(t, items, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, preHook)
	assert.Equal(t, 1, handler)
	assert.Equal(t, 2, postHook)
}

func TestResourceMultiGetPostHookOverwrite(t *testing.T) {
	i := NewIndex()
	s := newTestMStorer()
	s.multiGet = func(ctx context.Context, ids []interface{}) ([]*Item, error) {
		return []*Item{{ID: 1}}, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		*item = &Item{ID: 2}
	}))
	ctx := context.Background()
	items, err := r.MultiGet(ctx, []interface{}{1, 1})
	assert.Equal(t, []*Item{{ID: 2}, {ID: 2}}, items)
	assert.NoError(t, err)
}

func TestResourceMultiGetError(t *testing.T) {
	var preHook, postHook, handler int
	i := NewIndex()
	s := newTestMStorer()
	s.multiGet = func(ctx context.Context, ids []interface{}) ([]*Item, error) {
		handler++
		return nil, errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		preHook++
		assert.Equal(t, 1, id)
		return nil
	}))
	r.Use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		postHook++
		assert.Nil(t, *item)
		assert.EqualError(t, *err, "storer error")
	}))
	ctx := context.Background()
	items, err := r.MultiGet(ctx, []interface{}{1, 1})
	assert.Len(t, items, 0)
	assert.EqualError(t, err, "storer error")
	assert.Equal(t, 2, preHook)
	assert.Equal(t, 1, handler)
	assert.Equal(t, 2, postHook)
}

func TestResourceMultiGetPreHookError(t *testing.T) {
	var preHook, postHook, handler int
	i := NewIndex()
	s := newTestMStorer()
	s.multiGet = func(ctx context.Context, ids []interface{}) ([]*Item, error) {
		handler++
		return nil, errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		preHook++
		return errors.New("pre hook error")
	}))
	r.Use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		postHook++
		assert.Nil(t, *item)
		assert.EqualError(t, *err, "pre hook error")
	}))
	ctx := context.Background()
	items, err := r.MultiGet(ctx, []interface{}{1, 1})
	assert.Len(t, items, 0)
	assert.EqualError(t, err, "pre hook error")
	assert.Equal(t, 2, preHook)
	assert.Equal(t, 0, handler)
	assert.Equal(t, 2, postHook)
}

func TestResourceMultiGetPostHookError(t *testing.T) {
	var preHook, postHook, handler int
	i := NewIndex()
	s := newTestMStorer()
	s.multiGet = func(ctx context.Context, ids []interface{}) ([]*Item, error) {
		handler++
		return []*Item{{ID: 1}}, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(GetEventHandlerFunc(func(ctx context.Context, id interface{}) error {
		preHook++
		assert.Equal(t, 1, id)
		return nil
	}))
	r.Use(GotEventHandlerFunc(func(ctx context.Context, item **Item, err *error) {
		postHook++
		assert.Equal(t, &Item{ID: 1}, *item)
		assert.NoError(t, *err)
		*err = errors.New("post hook error")
	}))
	ctx := context.Background()
	items, err := r.MultiGet(ctx, []interface{}{1, 1})
	assert.Len(t, items, 0)
	assert.EqualError(t, err, "post hook error")
	assert.Equal(t, 2, preHook)
	assert.Equal(t, 1, handler)
	assert.Equal(t, 2, postHook)
}

/*
 * Find
 */

func TestResourceFind(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.find = func(ctx context.Context, lookup *Lookup, offset, limit int) (*ItemList, error) {
		handler = true
		return &ItemList{Items: []*Item{{ID: 1}}}, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(FindEventHandlerFunc(func(ctx context.Context, lookup *Lookup, offset, limit int) error {
		preHook = true
		assert.NotNil(t, lookup)
		assert.Equal(t, 0, offset)
		assert.Equal(t, 2, limit)
		return nil
	}))
	r.Use(FoundEventHandlerFunc(func(ctx context.Context, lookup *Lookup, list **ItemList, err *error) {
		postHook = true
		assert.NotNil(t, lookup)
		assert.Equal(t, &ItemList{Items: []*Item{{ID: 1}}}, *list)
		assert.NoError(t, *err)
	}))
	ctx := context.Background()
	_, err := r.Find(ctx, NewLookup(), 0, 2)
	assert.NoError(t, err)
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceMultiFindPostHookOverwrite(t *testing.T) {
	i := NewIndex()
	s := newTestMStorer()
	s.find = func(ctx context.Context, lookup *Lookup, offset, limit int) (*ItemList, error) {
		return &ItemList{Items: []*Item{{ID: 1}}}, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(FoundEventHandlerFunc(func(ctx context.Context, lookup *Lookup, list **ItemList, err *error) {
		*list = &ItemList{Items: []*Item{{ID: 2}}}
	}))
	ctx := context.Background()
	list, err := r.Find(ctx, NewLookup(), 0, 2)
	assert.Equal(t, &ItemList{Items: []*Item{{ID: 2}}}, list)
	assert.NoError(t, err)
}

func TestResourceFindError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.find = func(ctx context.Context, lookup *Lookup, offset, limit int) (*ItemList, error) {
		handler = true
		return nil, errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(FindEventHandlerFunc(func(ctx context.Context, lookup *Lookup, offset, limit int) error {
		preHook = true
		assert.NotNil(t, lookup)
		assert.Equal(t, 0, offset)
		assert.Equal(t, 2, limit)
		return nil
	}))
	r.Use(FoundEventHandlerFunc(func(ctx context.Context, lookup *Lookup, list **ItemList, err *error) {
		postHook = true
		assert.NotNil(t, lookup)
		assert.Nil(t, *list)
		assert.EqualError(t, *err, "storer error")
	}))
	ctx := context.Background()
	_, err := r.Find(ctx, NewLookup(), 0, 2)
	assert.EqualError(t, err, "storer error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceFindPreHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.find = func(ctx context.Context, lookup *Lookup, offset, limit int) (*ItemList, error) {
		handler = true
		return nil, errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(FindEventHandlerFunc(func(ctx context.Context, lookup *Lookup, offset, limit int) error {
		preHook = true
		return errors.New("pre hook error")
	}))
	r.Use(FoundEventHandlerFunc(func(ctx context.Context, lookup *Lookup, list **ItemList, err *error) {
		postHook = true
		assert.NotNil(t, lookup)
		assert.Nil(t, *list)
		assert.EqualError(t, *err, "pre hook error")
	}))
	ctx := context.Background()
	_, err := r.Find(ctx, NewLookup(), 0, 2)
	assert.EqualError(t, err, "pre hook error")
	assert.True(t, preHook)
	assert.False(t, handler)
	assert.True(t, postHook)
}

func TestResourceFindPostHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.find = func(ctx context.Context, lookup *Lookup, offset, limit int) (*ItemList, error) {
		handler = true
		return &ItemList{Items: []*Item{{ID: 1}}}, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(FindEventHandlerFunc(func(ctx context.Context, lookup *Lookup, offset, limit int) error {
		preHook = true
		assert.NotNil(t, lookup)
		assert.Equal(t, 0, offset)
		assert.Equal(t, 2, limit)
		return nil
	}))
	r.Use(FoundEventHandlerFunc(func(ctx context.Context, lookup *Lookup, list **ItemList, err *error) {
		postHook = true
		assert.NotNil(t, lookup)
		assert.Equal(t, &ItemList{Items: []*Item{{ID: 1}}}, *list)
		assert.NoError(t, *err)
		*err = errors.New("post hook error")
	}))
	ctx := context.Background()
	_, err := r.Find(ctx, NewLookup(), 0, 2)
	assert.EqualError(t, err, "post hook error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

/*
 * Insert
 */

func TestResourceInsert(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.insert = func(ctx context.Context, items []*Item) error {
		handler = true
		return nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(InsertEventHandlerFunc(func(ctx context.Context, items []*Item) error {
		preHook = true
		assert.Equal(t, []*Item{{ID: 1}}, items)
		return nil
	}))
	r.Use(InsertedEventHandlerFunc(func(ctx context.Context, items []*Item, err *error) {
		postHook = true
		assert.Equal(t, []*Item{{ID: 1}}, items)
		assert.NoError(t, *err)
	}))
	ctx := context.Background()
	err := r.Insert(ctx, []*Item{{ID: 1}})
	assert.NoError(t, err)
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceInsertError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.insert = func(ctx context.Context, items []*Item) error {
		handler = true
		return errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(InsertEventHandlerFunc(func(ctx context.Context, items []*Item) error {
		preHook = true
		assert.Equal(t, []*Item{{ID: 1}}, items)
		return nil
	}))
	r.Use(InsertedEventHandlerFunc(func(ctx context.Context, items []*Item, err *error) {
		postHook = true
		assert.Equal(t, []*Item{{ID: 1}}, items)
		assert.EqualError(t, *err, "storer error")
	}))
	ctx := context.Background()
	err := r.Insert(ctx, []*Item{{ID: 1}})
	assert.EqualError(t, err, "storer error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceInsertPreHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.insert = func(ctx context.Context, items []*Item) error {
		handler = true
		return errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(InsertEventHandlerFunc(func(ctx context.Context, items []*Item) error {
		preHook = true
		return errors.New("pre hook error")
	}))
	r.Use(InsertedEventHandlerFunc(func(ctx context.Context, items []*Item, err *error) {
		postHook = true
		assert.Equal(t, []*Item{{ID: 1}}, items)
		assert.EqualError(t, *err, "pre hook error")
	}))
	ctx := context.Background()
	err := r.Insert(ctx, []*Item{{ID: 1}})
	assert.EqualError(t, err, "pre hook error")
	assert.True(t, preHook)
	assert.False(t, handler)
	assert.True(t, postHook)
}

func TestResourceInsertPostHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.insert = func(ctx context.Context, items []*Item) error {
		handler = true
		return nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(InsertEventHandlerFunc(func(ctx context.Context, items []*Item) error {
		preHook = true
		assert.Equal(t, []*Item{{ID: 1}}, items)
		return nil
	}))
	r.Use(InsertedEventHandlerFunc(func(ctx context.Context, items []*Item, err *error) {
		postHook = true
		assert.Equal(t, []*Item{{ID: 1}}, items)
		assert.NoError(t, *err)
		*err = errors.New("post hook error")
	}))
	ctx := context.Background()
	err := r.Insert(ctx, []*Item{{ID: 1}})
	assert.EqualError(t, err, "post hook error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

/*
 * Update
 */

func TestResourceUpdate(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.update = func(ctx context.Context, item *Item, origin *Item) error {
		handler = true
		return nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(UpdateEventHandlerFunc(func(ctx context.Context, item *Item, origin *Item) error {
		preHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.Equal(t, &Item{ID: 1}, origin)
		return nil
	}))
	r.Use(UpdatedEventHandlerFunc(func(ctx context.Context, item *Item, origin *Item, err *error) {
		postHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.Equal(t, &Item{ID: 1}, origin)
		assert.NoError(t, *err)
	}))
	ctx := context.Background()
	err := r.Update(ctx, &Item{ID: 1}, &Item{ID: 1})
	assert.NoError(t, err)
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceUpdateError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.update = func(ctx context.Context, item *Item, origin *Item) error {
		handler = true
		return errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(UpdateEventHandlerFunc(func(ctx context.Context, item *Item, origin *Item) error {
		preHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.Equal(t, &Item{ID: 1}, origin)
		return nil
	}))
	r.Use(UpdatedEventHandlerFunc(func(ctx context.Context, item *Item, origin *Item, err *error) {
		postHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.Equal(t, &Item{ID: 1}, origin)
		assert.EqualError(t, *err, "storer error")
	}))
	ctx := context.Background()
	err := r.Update(ctx, &Item{ID: 1}, &Item{ID: 1})
	assert.EqualError(t, err, "storer error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceUpdatePreHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.update = func(ctx context.Context, item *Item, origin *Item) error {
		handler = true
		return errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(UpdateEventHandlerFunc(func(ctx context.Context, item *Item, origin *Item) error {
		preHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.Equal(t, &Item{ID: 1}, origin)
		return errors.New("pre hook error")
	}))
	r.Use(UpdatedEventHandlerFunc(func(ctx context.Context, item *Item, origin *Item, err *error) {
		postHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.Equal(t, &Item{ID: 1}, origin)
		assert.EqualError(t, *err, "pre hook error")
	}))
	ctx := context.Background()
	err := r.Update(ctx, &Item{ID: 1}, &Item{ID: 1})
	assert.EqualError(t, err, "pre hook error")
	assert.True(t, preHook)
	assert.False(t, handler)
	assert.True(t, postHook)
}

func TestResourceUpdatePostHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.update = func(ctx context.Context, item *Item, origin *Item) error {
		handler = true
		return nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(UpdateEventHandlerFunc(func(ctx context.Context, item *Item, origin *Item) error {
		preHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.Equal(t, &Item{ID: 1}, origin)
		return nil
	}))
	r.Use(UpdatedEventHandlerFunc(func(ctx context.Context, item *Item, origin *Item, err *error) {
		postHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.Equal(t, &Item{ID: 1}, origin)
		assert.NoError(t, *err)
		*err = errors.New("post hook error")
	}))
	ctx := context.Background()
	err := r.Update(ctx, &Item{ID: 1}, &Item{ID: 1})
	assert.EqualError(t, err, "post hook error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

/*
 * Delete
 */

func TestResourceDelete(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.delete = func(ctx context.Context, item *Item) error {
		handler = true
		return nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(DeleteEventHandlerFunc(func(ctx context.Context, item *Item) error {
		preHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		return nil
	}))
	r.Use(DeletedEventHandlerFunc(func(ctx context.Context, item *Item, err *error) {
		postHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.NoError(t, *err)
	}))
	ctx := context.Background()
	err := r.Delete(ctx, &Item{ID: 1})
	assert.NoError(t, err)
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceDeleteError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.delete = func(ctx context.Context, item *Item) error {
		handler = true
		return errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(DeleteEventHandlerFunc(func(ctx context.Context, item *Item) error {
		preHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		return nil
	}))
	r.Use(DeletedEventHandlerFunc(func(ctx context.Context, item *Item, err *error) {
		postHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.EqualError(t, *err, "storer error")
	}))
	ctx := context.Background()
	err := r.Delete(ctx, &Item{ID: 1})
	assert.EqualError(t, err, "storer error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceDeletePreHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.delete = func(ctx context.Context, item *Item) error {
		handler = true
		return errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(DeleteEventHandlerFunc(func(ctx context.Context, item *Item) error {
		preHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		return errors.New("pre hook error")
	}))
	r.Use(DeletedEventHandlerFunc(func(ctx context.Context, item *Item, err *error) {
		postHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.EqualError(t, *err, "pre hook error")
	}))
	ctx := context.Background()
	err := r.Delete(ctx, &Item{ID: 1})
	assert.EqualError(t, err, "pre hook error")
	assert.True(t, preHook)
	assert.False(t, handler)
	assert.True(t, postHook)
}

func TestResourceDeletePostHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.delete = func(ctx context.Context, item *Item) error {
		handler = true
		return nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(DeleteEventHandlerFunc(func(ctx context.Context, item *Item) error {
		preHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		return nil
	}))
	r.Use(DeletedEventHandlerFunc(func(ctx context.Context, item *Item, err *error) {
		postHook = true
		assert.Equal(t, &Item{ID: 1}, item)
		assert.NoError(t, *err)
		*err = errors.New("post hook error")
	}))
	ctx := context.Background()
	err := r.Delete(ctx, &Item{ID: 1})
	assert.EqualError(t, err, "post hook error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

/*
 * Clear
 */

func TestResourceClear(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.clear = func(ctx context.Context, lookup *Lookup) (int, error) {
		handler = true
		return 0, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(ClearEventHandlerFunc(func(ctx context.Context, lookup *Lookup) error {
		preHook = true
		assert.NotNil(t, lookup)
		return nil
	}))
	r.Use(ClearedEventHandlerFunc(func(ctx context.Context, lookup *Lookup, deleted *int, err *error) {
		postHook = true
		assert.NotNil(t, lookup)
		assert.NoError(t, *err)
	}))
	ctx := context.Background()
	_, err := r.Clear(ctx, NewLookup())
	assert.NoError(t, err)
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceClearError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.clear = func(ctx context.Context, lookup *Lookup) (int, error) {
		handler = true
		return 0, errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(ClearEventHandlerFunc(func(ctx context.Context, lookup *Lookup) error {
		preHook = true
		assert.NotNil(t, lookup)
		return nil
	}))
	r.Use(ClearedEventHandlerFunc(func(ctx context.Context, lookup *Lookup, deleted *int, err *error) {
		postHook = true
		assert.NotNil(t, lookup)
		assert.EqualError(t, *err, "storer error")
	}))
	ctx := context.Background()
	_, err := r.Clear(ctx, NewLookup())
	assert.EqualError(t, err, "storer error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}

func TestResourceClearPreHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.clear = func(ctx context.Context, lookup *Lookup) (int, error) {
		handler = true
		return 0, errors.New("storer error")
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(ClearEventHandlerFunc(func(ctx context.Context, lookup *Lookup) error {
		preHook = true
		assert.NotNil(t, lookup)
		return errors.New("pre hook error")
	}))
	r.Use(ClearedEventHandlerFunc(func(ctx context.Context, lookup *Lookup, deleted *int, err *error) {
		postHook = true
		assert.NotNil(t, lookup)
		assert.EqualError(t, *err, "pre hook error")
	}))
	ctx := context.Background()
	_, err := r.Clear(ctx, NewLookup())
	assert.EqualError(t, err, "pre hook error")
	assert.True(t, preHook)
	assert.False(t, handler)
	assert.True(t, postHook)
}

func TestResourceClearPostHookError(t *testing.T) {
	var preHook, postHook, handler bool
	i := NewIndex()
	s := newTestMStorer()
	s.clear = func(ctx context.Context, lookup *Lookup) (int, error) {
		handler = true
		return 0, nil
	}
	r := i.Bind("foo", schema.Schema{}, s, DefaultConf)
	r.Use(ClearEventHandlerFunc(func(ctx context.Context, lookup *Lookup) error {
		preHook = true
		assert.NotNil(t, lookup)
		return nil
	}))
	r.Use(ClearedEventHandlerFunc(func(ctx context.Context, lookup *Lookup, deleted *int, err *error) {
		postHook = true
		assert.NotNil(t, lookup)
		assert.NoError(t, *err)
		*err = errors.New("post hook error")
	}))
	ctx := context.Background()
	_, err := r.Clear(ctx, NewLookup())
	assert.EqualError(t, err, "post hook error")
	assert.True(t, preHook)
	assert.True(t, handler)
	assert.True(t, postHook)
}
