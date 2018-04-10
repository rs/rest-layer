// Package mem is an example REST backend storage that stores everything in memory.
package mem

import (
	"bytes"
	"context"
	"encoding/gob"
	"sort"
	"sync"
	"time"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema/query"
)

// MemoryHandler is an example handler storing data in memory.
type MemoryHandler struct {
	sync.RWMutex

	// If Latency is set, the handler will introduce an artificial latency on
	// all operations.
	Latency time.Duration

	items map[interface{}][]byte
	ids   []interface{}
}

func init() {
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
	gob.Register(time.Time{})
}

// NewHandler creates an empty memory handler.
func NewHandler() *MemoryHandler {
	return &MemoryHandler{
		items: map[interface{}][]byte{},
		ids:   []interface{}{},
	}
}

// NewSlowHandler creates an empty memory handler with specified latency.
func NewSlowHandler(latency time.Duration) *MemoryHandler {
	return &MemoryHandler{
		Latency: latency,
		items:   map[interface{}][]byte{},
		ids:     []interface{}{},
	}
}

// store serialize the item using gob and store it in the handler's items map.
func (m *MemoryHandler) store(item *resource.Item) error {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	if err := enc.Encode(*item); err != nil {
		return err
	}
	m.items[item.ID] = data.Bytes()
	return nil
}

// fetch finds an item by ID and returns a new item with unserialized data.
func (m *MemoryHandler) fetch(id interface{}) (*resource.Item, bool, error) {
	data, found := m.items[id]
	if !found {
		return nil, false, nil
	}
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	var item resource.Item
	if err := dec.Decode(&item); err != nil {
		return nil, true, err
	}
	return &item, true, nil
}

// delete removes an item by this id without locking.
func (m *MemoryHandler) delete(id interface{}) {
	delete(m.items, id)
	// Remove id from id list
	for i, _id := range m.ids {
		if _id == id {
			if i >= len(m.ids)-1 {
				m.ids = m.ids[:i]
			} else {
				m.ids = append(m.ids[:i], m.ids[i+1:]...)
			}
			break
		}
	}
}

// Insert inserts new items in memory.
func (m *MemoryHandler) Insert(ctx context.Context, items []*resource.Item) (err error) {
	m.Lock()
	defer m.Unlock()
	err = handleWithLatency(m.Latency, ctx, func() error {
		for _, item := range items {
			if _, found := m.items[item.ID]; found {
				return resource.ErrConflict
			}
		}
		for _, item := range items {
			if err := m.store(item); err != nil {
				return err
			}
			// Store ids in ordered slice for sorting
			m.ids = append(m.ids, item.ID)
		}
		return nil
	})
	return err
}

// Update replace an item by a new one in memory.
func (m *MemoryHandler) Update(ctx context.Context, item *resource.Item, original *resource.Item) (err error) {
	m.Lock()
	defer m.Unlock()
	err = handleWithLatency(m.Latency, ctx, func() error {
		o, found, err := m.fetch(original.ID)
		if !found {
			return resource.ErrNotFound
		}
		if err != nil {
			return err
		}
		if original.ETag != o.ETag {
			return resource.ErrConflict
		}
		return m.store(item)
	})
	return err
}

// Delete deletes an item from memory.
func (m *MemoryHandler) Delete(ctx context.Context, item *resource.Item) (err error) {
	m.Lock()
	defer m.Unlock()
	err = handleWithLatency(m.Latency, ctx, func() error {
		o, found, err := m.fetch(item.ID)
		if !found {
			return resource.ErrNotFound
		}
		if err != nil {
			return err
		}
		if item.ETag != o.ETag {
			return resource.ErrConflict
		}
		m.delete(item.ID)
		return nil
	})
	return err
}

// Clear clears all items from the memory store matching q.
func (m *MemoryHandler) Clear(ctx context.Context, q *query.Query) (total int, err error) {
	m.Lock()
	defer m.Unlock()
	err = handleWithLatency(m.Latency, ctx, func() error {
		list, err := m.find(ctx, q)
		if err != nil {
			return err
		}
		for _, item := range list.Items {
			m.delete(item.ID)
			total++
		}
		return nil
	})
	return total, err
}

// Find items from memory matching the q.
func (m *MemoryHandler) Find(ctx context.Context, q *query.Query) (list *resource.ItemList, err error) {
	m.RLock()
	defer m.RUnlock()

	err = handleWithLatency(m.Latency, ctx, func() error {
		list, err = m.find(ctx, q)
		return err
	})
	return list, err
}

func (m *MemoryHandler) find(ctx context.Context, q *query.Query) (*resource.ItemList, error) {
	// Fetch all items matching the filter
	list := resource.ItemList{Items: []*resource.Item{}}
	for _, id := range m.ids {
		item, _, err := m.fetch(id)
		if err != nil {
			return nil, err
		}
		if !q.Predicate.Match(item.Payload) {
			continue
		}
		list.Items = append(list.Items, item)
	}
	list.Total = len(list.Items)

	// Apply sort
	if len(q.Sort) > 0 {
		s := sortableItems{q.Sort, list.Items}
		sort.Sort(s)
	}
	// Apply pagination
	if q.Window != nil {
		list.Limit = q.Window.Limit
		list.Offset = q.Window.Offset

		if list.Offset > list.Total-1 {
			list.Items = nil
		} else {
			list.Items = list.Items[list.Offset:]
			if list.Limit >= 0 && list.Limit < len(list.Items) {
				list.Items = list.Items[:list.Limit]
			}
		}

	}
	return &list, nil
}
