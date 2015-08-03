package rest

import (
	"net/url"
	"testing"

	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// mockHandler is a read-only storage handler which always return what is stored in items
// with no support for filtering/sorting or error if err is set
type mockHandler struct {
	items []*Item
	err   *Error
}

func (m *mockHandler) Insert(items []*Item, ctx context.Context) *Error {
	return NotImplementedError
}
func (m *mockHandler) Update(item *Item, original *Item, ctx context.Context) *Error {
	return NotImplementedError
}
func (m *mockHandler) Delete(item *Item, ctx context.Context) *Error {
	return NotImplementedError
}
func (m *mockHandler) Clear(lookup *Lookup, ctx context.Context) (int, *Error) {
	return 0, NotImplementedError
}
func (m *mockHandler) Find(lookup *Lookup, page, perPage int, ctx context.Context) (*ItemList, *Error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ItemList{len(m.items), page, m.items}, nil
}

func newRoute(method string) route {
	return route{
		method: method,
		fields: map[string]interface{}{},
		params: url.Values{},
	}
}

func TestFindRoute(t *testing.T) {
	var route route
	var err *Error
	root := New()
	i, _ := NewItem(map[string]interface{}{"id": "1234"})
	h := &mockHandler{[]*Item{i}, nil}
	foo := root.Bind("foo", NewResource(schema.Schema{}, h, DefaultConf))
	bar := foo.Bind("bar", "foo", NewResource(schema.Schema{"foo": schema.Field{}}, h, DefaultConf))
	bar.Alias("baz", url.Values{"sort": []string{"foo"}})
	ctx := context.Background()

	route = newRoute("GET")
	err = findRoute(ctx, "/foo", root.resources, &route)
	assert.Nil(t, err)
	assert.Equal(t, foo, route.resource)
	assert.Equal(t, url.Values{}, route.params)
	assert.Nil(t, route.fields["id"])

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234", root.resources, &route)
	assert.Nil(t, err)
	assert.Equal(t, foo, route.resource)
	assert.Equal(t, url.Values{}, route.params)
	assert.Equal(t, "1234", route.fields["id"])

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234/bar", root.resources, &route)
	assert.Nil(t, err)
	assert.Equal(t, bar, route.resource)
	assert.Equal(t, "1234", route.fields["foo"])
	assert.Equal(t, url.Values{}, route.params)
	assert.Nil(t, route.fields["id"])

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234/bar/baz", root.resources, &route)
	assert.Nil(t, err)
	assert.Equal(t, bar, route.resource)
	assert.Equal(t, "1234", route.fields["foo"])
	assert.Equal(t, url.Values{"sort": []string{"foo"}}, route.params)
	assert.Nil(t, route.fields["id"])

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234/bar/baz/baz", root.resources, &route)
	assert.Equal(t, &Error{404, "Resource Not Found", nil}, err)

	route = newRoute("GET")
	// empty the storage handler
	h.items = []*Item{}
	err = findRoute(ctx, "/foo/1234/bar", root.resources, &route)
	assert.Equal(t, NotFoundError, err)

	route = newRoute("GET")
	// for error
	h.err = &Error{123, "", nil}
	err = findRoute(ctx, "/foo/1234/bar", root.resources, &route)
	assert.Equal(t, h.err, err)
}

func TestRouteApplyFields(t *testing.T) {
	r := route{
		fields: map[string]interface{}{
			"id":   "123",
			"user": "john",
		},
	}
	p := map[string]interface{}{"id": "321", "name": "John Doe"}
	r.applyFields(p)
	assert.Equal(t, map[string]interface{}{"id": "123", "user": "john", "name": "John Doe"}, p)
}
