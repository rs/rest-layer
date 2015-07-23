package rest

import (
	"log"
	"net/url"

	"github.com/rs/rest-layer/schema"
)

// Resource holds information about a class of items exposed on the API
type Resource struct {
	schema    schema.Validator
	handler   ResourceHandler
	conf      Conf
	resources map[string]*subResource
	aliases   map[string]url.Values
}

// subResource is used to bind resources and sub-resources
type subResource struct {
	field    string
	resource *Resource
}

// ResourceHandler defines the interface of an handler able to manage the life of a resource
type ResourceHandler interface {
	// Find searches for items in the backend store matching the lookup argument. The
	// pagination argument must be respected. If no items are found, an empty list
	// should be returned with no error.
	//
	// If the total number of item can't be easily computed, ItemList.Total should
	// be set to -1. The requested page should be set to ItemList.Page.
	Find(lookup *Lookup, page, perPage int) (*ItemList, *Error)
	// Store stores an item to the backend store. If the original item is provided, the
	// ResourceHandler must ensure that the item exists in the database and has the same
	// Etag field. This check should be performed atomically. If the original item is not
	// found, a rest.NotFoundError must be returned. If the etags don't match, a
	// rest.ConflictError must be returned.
	//
	// The item payload must be stored together with the etag and the updated field using.
	// The item.ID and the payload["id"] is garantied to be identical, so there's not need
	// to store both.
	Store(item *Item, original *Item) *Error
	// Delete deletes the provided item by its ID. The Etag of the item stored in the
	// backend store must match the Etag of the provided item or a rest.ConflictError
	// must be returned. This check should be performed atomically.
	//
	// If the provided item were not present in the backend store, a rest.NotFoundError
	// must be returned.
	Delete(item *Item) *Error
	// Clear removes all items maching the lookup. When possible, the number of items
	// removed is returned, otherwise -1 is return as the first value.
	Clear(lookup *Lookup) (int, *Error)
}

// NewResource creates a new resource with provided spec, handler and config
func NewResource(s schema.Validator, h ResourceHandler, c Conf) *Resource {
	return &Resource{
		schema:    s,
		handler:   h,
		conf:      c,
		resources: map[string]*subResource{},
		aliases:   map[string]url.Values{},
	}
}

// Bind a sub-resource with the provided name. The field parameter defines the parent
// resource's which contains the sub resource id.
//
//     users := api.Bind("users", userResource)
//     // Bind a sub resource on /users/:user_id/posts[/:post_id]
//     // and reference the user on each post using the "user" field.
//     posts := users.Bind("posts", "user", postResource)
//
// This method will panic an alias or a resource with the same name is already bound
// or if the specified field doesn't exist in the parent resource spec.
func (r *Resource) Bind(name, field string, s *Resource) *Resource {
	r.assertNotBound(name)
	if f := s.schema.GetField(field); f == nil {
		log.Panicf("Cannot bind `%s' as sub-resource: field `%s' does not exist in the sub-resource'", name, field)
	}
	// Compile schema and panic on any compilation error
	if c, ok := s.schema.(schema.Compiler); ok {
		if err := c.Compile(); err != nil {
			log.Fatalf("Schema compilation error: %s.%s", name, err)
		}
	}
	r.resources[name] = &subResource{
		field:    field,
		resource: s,
	}
	return s
}

// Alias adds an pre-built resource query on /<resource>/<alias>.
//
//     // Add a friendly alias to public posts on /users/:user_id/posts/public
//     // (equivalent to /users/:user_id/posts?filter=public=true)
//     posts.Alias("public", url.Values{"where": []string{"public=true"}})
//
// This method will panic an alias or a resource with the same name is already bound
func (r *Resource) Alias(name string, v url.Values) {
	r.assertNotBound(name)
	r.aliases[name] = v
}

// assertNotBound asserts a given resource name is not already bound
func (r *Resource) assertNotBound(name string) {
	if _, found := r.resources[name]; found {
		log.Panicf("Cannot bind `%s': already bound as resource'", name)
	}
	if _, found := r.aliases[name]; found {
		log.Panicf("Cannot bind `%s': already bound as alias'", name)
	}
}
