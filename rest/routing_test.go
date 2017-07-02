package rest

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"testing"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
	"github.com/stretchr/testify/assert"
)

// mockHandler is a read-only storage handler which always return what is stored in items
// with no support for filtering/sorting or error if err is set
type mockHandler struct {
	items   []*resource.Item
	err     error
	queries []query.Query
	lock    sync.Mutex
}

func (m *mockHandler) Insert(ctx context.Context, items []*resource.Item) error {
	return ErrNotImplemented
}
func (m *mockHandler) Update(ctx context.Context, item *resource.Item, original *resource.Item) error {
	return ErrNotImplemented
}
func (m *mockHandler) Delete(ctx context.Context, item *resource.Item) error {
	return ErrNotImplemented
}
func (m *mockHandler) Clear(ctx context.Context, lookup *resource.Lookup) (int, error) {
	return 0, ErrNotImplemented
}
func (m *mockHandler) Find(ctx context.Context, lookup *resource.Lookup, offset, limit int) (*resource.ItemList, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	m.queries = append(m.queries, lookup.Filter())
	return &resource.ItemList{Total: len(m.items), Items: m.items}, nil
}

func newRoute(method string) *RouteMatch {
	return &RouteMatch{
		Method:       method,
		ResourcePath: ResourcePath{},
		Params:       url.Values{},
	}
}

func TestFindRoute(t *testing.T) {
	var route *RouteMatch
	var err error
	index := resource.NewIndex()
	i, _ := resource.NewItem(map[string]interface{}{"id": "1234"})
	h := &mockHandler{[]*resource.Item{i}, nil, []query.Query{}, sync.Mutex{}}
	foo := index.Bind("foo", schema.Schema{}, h, resource.DefaultConf)
	bar := foo.Bind("bar", "f", schema.Schema{Fields: schema.Fields{"f": {}}}, h, resource.DefaultConf)
	barbar := bar.Bind("bar", "b", schema.Schema{Fields: schema.Fields{"b": {}}}, h, resource.DefaultConf)
	bar.Alias("baz", url.Values{"sort": []string{"foo"}})

	route = newRoute("GET")
	err = findRoute("/foo", index, route)
	if assert.Nil(t, err) {
		assert.Equal(t, foo, route.Resource())
		assert.Equal(t, url.Values{}, route.Params)
		assert.Nil(t, route.ResourceID())
		rp := route.ResourcePath
		if assert.Len(t, rp, 1) {
			assert.Equal(t, "foo", rp[0].Name)
			assert.Equal(t, "", rp[0].Field)
			assert.Nil(t, rp[0].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute("/foo/1234", index, route)
	if assert.Nil(t, err) {
		assert.Equal(t, foo, route.Resource())
		assert.Equal(t, url.Values{}, route.Params)
		assert.Equal(t, "1234", route.ResourceID())
		rp := route.ResourcePath
		if assert.Len(t, rp, 1) {
			assert.Equal(t, "foo", rp[0].Name)
			assert.Equal(t, "id", rp[0].Field)
			assert.Equal(t, "1234", rp[0].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute("/foo/1234/bar", index, route)
	if assert.Nil(t, err) {
		assert.Equal(t, bar, route.Resource())
		assert.Nil(t, route.ResourceID())
		assert.Equal(t, url.Values{}, route.Params)
		rp := route.ResourcePath
		if assert.Len(t, rp, 2) {
			assert.Equal(t, "foo", rp[0].Name)
			assert.Equal(t, "f", rp[0].Field)
			assert.Equal(t, "1234", rp[0].Value)
			assert.Equal(t, "bar", rp[1].Name)
			assert.Equal(t, "", rp[1].Field)
			assert.Nil(t, rp[1].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute("/foo/1234/bar/1234", index, route)
	if assert.Nil(t, err) {
		assert.Equal(t, bar, route.Resource())
		assert.Equal(t, "1234", route.ResourceID())
		assert.Equal(t, url.Values{}, route.Params)
		rp := route.ResourcePath
		if assert.Len(t, rp, 2) {
			assert.Equal(t, "foo", rp[0].Name)
			assert.Equal(t, "f", rp[0].Field)
			assert.Equal(t, "1234", rp[0].Value)
			assert.Equal(t, "bar", rp[1].Name)
			assert.Equal(t, "id", rp[1].Field)
			assert.Equal(t, "1234", rp[1].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute("/foo/1234/bar/1234/bar", index, route)
	if assert.Nil(t, err) {
		assert.Equal(t, barbar, route.Resource())
		assert.Nil(t, route.ResourceID())
		assert.Equal(t, url.Values{}, route.Params)
		rp := route.ResourcePath
		if assert.Len(t, rp, 3) {
			assert.Equal(t, "foo", rp[0].Name)
			assert.Equal(t, "f", rp[0].Field)
			assert.Equal(t, "1234", rp[0].Value)
			assert.Equal(t, "bar", rp[1].Name)
			assert.Equal(t, "b", rp[1].Field)
			assert.Equal(t, "1234", rp[1].Value)
			assert.Equal(t, "bar", rp[2].Name)
			assert.Equal(t, "", rp[2].Field)
			assert.Nil(t, rp[2].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute("/foo/1234/bar/1234/bar/1234", index, route)
	if assert.Nil(t, err) {
		assert.Equal(t, barbar, route.Resource())
		assert.Equal(t, "1234", route.ResourceID())
		assert.Equal(t, url.Values{}, route.Params)
		rp := route.ResourcePath
		if assert.Len(t, rp, 3) {
			assert.Equal(t, "foo", rp[0].Name)
			assert.Equal(t, "f", rp[0].Field)
			assert.Equal(t, "1234", rp[0].Value)
			assert.Equal(t, "bar", rp[1].Name)
			assert.Equal(t, "b", rp[1].Field)
			assert.Equal(t, "1234", rp[1].Value)
			assert.Equal(t, "bar", rp[2].Name)
			assert.Equal(t, "id", rp[2].Field)
			assert.Equal(t, "1234", rp[2].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute("/foo/1234/bar/baz", index, route)
	if assert.Nil(t, err) {
		assert.Equal(t, bar, route.Resource())
		assert.Equal(t, url.Values{"sort": []string{"foo"}}, route.Params)
		assert.Nil(t, route.ResourceID())
		rp := route.ResourcePath
		if assert.Len(t, rp, 2) {
			assert.Equal(t, "foo", rp[0].Name)
			assert.Equal(t, "f", rp[0].Field)
			assert.Equal(t, "1234", rp[0].Value)
			assert.Equal(t, "bar", rp[1].Name)
			assert.Equal(t, "", rp[1].Field)
			assert.Nil(t, rp[1].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute("/foo/1234/bar/baz/baz", index, route)
	assert.Equal(t, &Error{404, "Resource Not Found", nil}, err)
	assert.Nil(t, route.Resource())
	assert.Nil(t, route.ResourceID())
}

func TestRoutePathParentsExists(t *testing.T) {
	var route *RouteMatch
	var err error
	index := resource.NewIndex()
	i, _ := resource.NewItem(map[string]interface{}{"id": "1234"})
	h := &mockHandler{[]*resource.Item{i}, nil, []query.Query{}, sync.Mutex{}}
	foo := index.Bind("foo", schema.Schema{}, h, resource.DefaultConf)
	bar := foo.Bind("bar", "f", schema.Schema{Fields: schema.Fields{"f": {}}}, h, resource.DefaultConf)
	bar.Bind("baz", "b", schema.Schema{Fields: schema.Fields{"f": {}, "b": {}}}, h, resource.DefaultConf)
	ctx := context.Background()

	route = newRoute("GET")
	err = findRoute("/foo/1234/bar/5678/baz/9000", index, route)
	if assert.NoError(t, err) {
		err = route.ResourcePath.ParentsExist(ctx)
		assert.NoError(t, err)
		// There's 3 components in the path but only 2 are parents
		assert.Len(t, h.queries, 2)
		// query on /foo/1234
		assert.Contains(t, h.queries, query.Query{query.Equal{Field: "id", Value: "1234"}})
		// query on /bar/5678 with foo/1234 context
		assert.Contains(t, h.queries, query.Query{query.Equal{Field: "f", Value: "1234"}, query.Equal{Field: "id", Value: "5678"}})
	}

	route = newRoute("GET")
	// empty the storage handler
	h.items = []*resource.Item{}
	err = findRoute("/foo/1234/bar", index, route)
	if assert.NoError(t, err) {
		err = route.ResourcePath.ParentsExist(ctx)
		assert.Equal(t, &Error{404, "Parent Resource Not Found", nil}, err)
	}

	route = newRoute("GET")
	// for error
	h.err = errors.New("test")
	err = findRoute("/foo/1234/bar", index, route)
	if assert.NoError(t, err) {
		err = route.ResourcePath.ParentsExist(ctx)
		assert.EqualError(t, err, "test")
	}
}

func TestRoutePathParentsNotExists(t *testing.T) {
	index := resource.NewIndex()
	i, _ := resource.NewItem(map[string]interface{}{"id": "1234"})
	h := &mockHandler{[]*resource.Item{i}, nil, []query.Query{}, sync.Mutex{}}
	empty := &mockHandler{[]*resource.Item{}, nil, []query.Query{}, sync.Mutex{}}
	foo := index.Bind("foo", schema.Schema{}, empty, resource.DefaultConf)
	foo.Bind("bar", "f", schema.Schema{Fields: schema.Fields{"f": {}}}, h, resource.DefaultConf)
	ctx := context.Background()

	route := newRoute("GET")
	// non existing foo
	err := findRoute("/foo/4321/bar/1234", index, route)
	if assert.NoError(t, err) {
		err := route.ResourcePath.ParentsExist(ctx)
		assert.Equal(t, &Error{404, "Parent Resource Not Found", nil}, err)
	}
}
