package resource

import (
	"fmt"

	"github.com/afex/hystrix-go/hystrix"
	"golang.org/x/net/context"
)

type hystrixStorage struct {
	getCmd      string
	multiGetCmd string
	findCmd     string
	insertCmd   string
	updateCmd   string
	deleteCmd   string
	clearCmd    string
	storage     storageHandler
}

func newHystrixStorage(name string, s storageHandler) hystrixStorage {
	return hystrixStorage{
		getCmd:      fmt.Sprintf("%s.Get", name),
		multiGetCmd: fmt.Sprintf("%s.MultiGet", name),
		findCmd:     fmt.Sprintf("%s.Find", name),
		insertCmd:   fmt.Sprintf("%s.Insert", name),
		updateCmd:   fmt.Sprintf("%s.Update", name),
		deleteCmd:   fmt.Sprintf("%s.Delete", name),
		clearCmd:    fmt.Sprintf("%s.Clear", name),
		storage:     s,
	}
}

func (h hystrixStorage) Get(ctx context.Context, id interface{}) (item *Item, err error) {
	out := make(chan *Item, 1)
	errs := hystrix.Go(h.getCmd, func() error {
		item, err := h.storage.Get(ctx, []interface{}{id})
		if err == nil {
			out <- item
		}
		return err
	}, nil)
	select {
	case item = <-out:
	case err = <-errs:
	}
	return
}

func (h hystrixStorage) MultiGet(ctx context.Context, ids []interface{}) (items []*Item, err error) {
	out := make(chan []*Item, 1)
	errs := hystrix.Go(h.multiGetCmd, func() error {
		items, err := h.storage.MultiGet(ctx, ids)
		if err == nil {
			out <- items
		}
		return err
	}, nil)
	select {
	case items = <-out:
	case err = <-errs:
	}
	return
}

func (h hystrixStorage) Find(ctx context.Context, lookup *Lookup, page, perPage int) (list *ItemList, err error) {
	out := make(chan *ItemList, 1)
	errs := hystrix.Go(h.findCmd, func() error {
		list, err := h.storage.Find(ctx, lookup, page, perPage)
		if err == nil {
			out <- list
		}
		return err
	}, nil)
	select {
	case list = <-out:
	case err = <-errs:
	}
	return
}

func (h hystrixStorage) Insert(ctx context.Context, items []*Item) (err error) {
	return hystrix.Do(h.insertCmd, func() error {
		return h.storage.Insert(ctx, items)
	}, nil)
}

func (h hystrixStorage) Update(ctx context.Context, item *Item, original *Item) (err error) {
	return hystrix.Do(h.updateCmd, func() error {
		return h.storage.Update(ctx, item, original)
	}, nil)
}

func (h hystrixStorage) Delete(ctx context.Context, item *Item) (err error) {
	return hystrix.Do(h.deleteCmd, func() error {
		return h.storage.Delete(ctx, item)
	}, nil)
}

func (h hystrixStorage) Clear(ctx context.Context, lookup *Lookup) (deleted int, err error) {
	out := make(chan int, 1)
	errs := hystrix.Go(h.clearCmd, func() error {
		deleted, err := h.storage.Clear(ctx, lookup)
		if err == nil {
			out <- deleted
		}
		return err
	}, nil)
	select {
	case deleted = <-out:
	case err = <-errs:
	}
	return
}
