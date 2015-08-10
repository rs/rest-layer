package rest

import (
	"errors"
	"net/url"
	"testing"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// mockHandler is a read-only storage handler which always return what is stored in items
// with no support for filtering/sorting or error if err is set
type mockHandler struct {
	items []*resource.Item
	err   error
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
func (m *mockHandler) Find(ctx context.Context, lookup *resource.Lookup, page, perPage int) (*resource.ItemList, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &resource.ItemList{len(m.items), page, m.items}, nil
}

func newRoute(method string) RouteMatch {
	return RouteMatch{
		Method:       method,
		ResourcePath: []ResourcePathComponent{},
		Params:       url.Values{},
	}
}

func TestFindRoute(t *testing.T) {
	var route RouteMatch
	var err *Error
	index := resource.NewIndex()
	i, _ := resource.NewItem(map[string]interface{}{"id": "1234"})
	h := &mockHandler{[]*resource.Item{i}, nil}
	foo := index.Bind("foo", resource.New(schema.Schema{}, h, resource.DefaultConf))
	bar := foo.Bind("bar", "f", resource.New(schema.Schema{"f": schema.Field{}}, h, resource.DefaultConf))
	barbar := bar.Bind("bar", "b", resource.New(schema.Schema{"b": schema.Field{}}, h, resource.DefaultConf))
	bar.Alias("baz", url.Values{"sort": []string{"foo"}})
	ctx := context.Background()

	route = newRoute("GET")
	err = findRoute(ctx, "/foo", index, &route)
	if assert.Nil(t, err) {
		assert.Equal(t, foo, route.Resource())
		assert.Equal(t, url.Values{}, route.Params)
		assert.Nil(t, route.ResourceID())
		if assert.Len(t, route.ResourcePath, 1) {
			assert.Equal(t, "foo", route.ResourcePath[0].Name)
			assert.Equal(t, "", route.ResourcePath[0].Field)
			assert.Nil(t, route.ResourcePath[0].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234", index, &route)
	if assert.Nil(t, err) {
		assert.Equal(t, foo, route.Resource())
		assert.Equal(t, url.Values{}, route.Params)
		assert.Equal(t, "1234", route.ResourceID())
		if assert.Len(t, route.ResourcePath, 1) {
			assert.Equal(t, "foo", route.ResourcePath[0].Name)
			assert.Equal(t, "id", route.ResourcePath[0].Field)
			assert.Equal(t, "1234", route.ResourcePath[0].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234/bar", index, &route)
	if assert.Nil(t, err) {
		assert.Equal(t, bar, route.Resource())
		assert.Nil(t, route.ResourceID())
		assert.Equal(t, url.Values{}, route.Params)
		if assert.Len(t, route.ResourcePath, 2) {
			assert.Equal(t, "foo", route.ResourcePath[0].Name)
			assert.Equal(t, "f", route.ResourcePath[0].Field)
			assert.Equal(t, "1234", route.ResourcePath[0].Value)
			assert.Equal(t, "bar", route.ResourcePath[1].Name)
			assert.Equal(t, "", route.ResourcePath[1].Field)
			assert.Nil(t, route.ResourcePath[1].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234/bar/1234", index, &route)
	if assert.Nil(t, err) {
		assert.Equal(t, bar, route.Resource())
		assert.Equal(t, "1234", route.ResourceID())
		assert.Equal(t, url.Values{}, route.Params)
		if assert.Len(t, route.ResourcePath, 2) {
			assert.Equal(t, "foo", route.ResourcePath[0].Name)
			assert.Equal(t, "f", route.ResourcePath[0].Field)
			assert.Equal(t, "1234", route.ResourcePath[0].Value)
			assert.Equal(t, "bar", route.ResourcePath[1].Name)
			assert.Equal(t, "id", route.ResourcePath[1].Field)
			assert.Equal(t, "1234", route.ResourcePath[1].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234/bar/1234/bar", index, &route)
	if assert.Nil(t, err) {
		assert.Equal(t, barbar, route.Resource())
		assert.Nil(t, route.ResourceID())
		assert.Equal(t, url.Values{}, route.Params)
		if assert.Len(t, route.ResourcePath, 3) {
			assert.Equal(t, "foo", route.ResourcePath[0].Name)
			assert.Equal(t, "f", route.ResourcePath[0].Field)
			assert.Equal(t, "1234", route.ResourcePath[0].Value)
			assert.Equal(t, "bar", route.ResourcePath[1].Name)
			assert.Equal(t, "b", route.ResourcePath[1].Field)
			assert.Equal(t, "1234", route.ResourcePath[1].Value)
			assert.Equal(t, "bar", route.ResourcePath[2].Name)
			assert.Equal(t, "", route.ResourcePath[2].Field)
			assert.Nil(t, route.ResourcePath[2].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234/bar/1234/bar/1234", index, &route)
	if assert.Nil(t, err) {
		assert.Equal(t, barbar, route.Resource())
		assert.Equal(t, "1234", route.ResourceID())
		assert.Equal(t, url.Values{}, route.Params)
		if assert.Len(t, route.ResourcePath, 3) {
			assert.Equal(t, "foo", route.ResourcePath[0].Name)
			assert.Equal(t, "f", route.ResourcePath[0].Field)
			assert.Equal(t, "1234", route.ResourcePath[0].Value)
			assert.Equal(t, "bar", route.ResourcePath[1].Name)
			assert.Equal(t, "b", route.ResourcePath[1].Field)
			assert.Equal(t, "1234", route.ResourcePath[1].Value)
			assert.Equal(t, "bar", route.ResourcePath[2].Name)
			assert.Equal(t, "id", route.ResourcePath[2].Field)
			assert.Equal(t, "1234", route.ResourcePath[2].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234/bar/baz", index, &route)
	if assert.Nil(t, err) {
		assert.Equal(t, bar, route.Resource())
		assert.Equal(t, url.Values{"sort": []string{"foo"}}, route.Params)
		assert.Nil(t, route.ResourceID())
		if assert.Len(t, route.ResourcePath, 2) {
			assert.Equal(t, "foo", route.ResourcePath[0].Name)
			assert.Equal(t, "f", route.ResourcePath[0].Field)
			assert.Equal(t, "1234", route.ResourcePath[0].Value)
			assert.Equal(t, "bar", route.ResourcePath[1].Name)
			assert.Equal(t, "", route.ResourcePath[1].Field)
			assert.Nil(t, route.ResourcePath[1].Value)
		}
	}

	route = newRoute("GET")
	err = findRoute(ctx, "/foo/1234/bar/baz/baz", index, &route)
	assert.Equal(t, &Error{404, "Resource Not Found", nil}, err)
	assert.Nil(t, route.Resource())
	assert.Nil(t, route.ResourceID())

	route = newRoute("GET")
	// empty the storage handler
	h.items = []*resource.Item{}
	err = findRoute(ctx, "/foo/1234/bar", index, &route)
	assert.Equal(t, ErrNotFound, err)

	route = newRoute("GET")
	// for error
	h.err = errors.New("test")
	err = findRoute(ctx, "/foo/1234/bar", index, &route)
	assert.Equal(t, &Error{520, "test", nil}, err)
}

func TestRouteApplyFields(t *testing.T) {
	r := RouteMatch{
		ResourcePath: ResourcePath{
			ResourcePathComponent{
				Name:  "users",
				Field: "user",
				Value: "john",
			},
			ResourcePathComponent{
				Name:  "posts",
				Field: "id",
				Value: "123",
			},
		},
	}
	p := map[string]interface{}{"id": "321", "name": "John Doe"}
	r.applyFields(p)
	assert.Equal(t, map[string]interface{}{"id": "123", "user": "john", "name": "John Doe"}, p)
}
