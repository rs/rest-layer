package resource

import (
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/rs/rest-layer/schema"
	"golang.org/x/net/context"
)

// Resource holds information about a class of items exposed on the API
type Resource struct {
	validator     validatorFallback
	storage       Storer
	conf          Conf
	resources     map[string]*subResource
	lockResources sync.Mutex
	aliases       map[string]url.Values
}

// subResource is used to bind resources and sub-resources
type subResource struct {
	field    string
	resource *Resource
}

// validatorFallback wraps a validator and fallback on given schema if the GetField
// returns nil on a given name
type validatorFallback struct {
	schema.Validator
	schema schema.Schema
}

func (v validatorFallback) GetField(name string) *schema.Field {
	if f := v.Validator.GetField(name); f != nil {
		return f
	}
	return v.schema.GetField(name)
}

// connection is a special internal validator to hook a validator of a sub resource
// to the resource validator in order to allow embedding of sub resources during
// field selection. Those connections are set on a fallback schema.
type connection struct {
	path string
}

func (v connection) Validate(value interface{}) (interface{}, error) {
	// no validation needed
	return value, nil
}

// New creates a new resource with provided spec, handler and config
func New(v schema.Validator, s Storer, c Conf) *Resource {
	return &Resource{
		validator: validatorFallback{
			Validator: v,
			schema:    schema.Schema{},
		},
		storage:   s,
		conf:      c,
		resources: map[string]*subResource{},
		aliases:   map[string]url.Values{},
	}
}

// Compile the resource graph and report any error
func (r *Resource) Compile() error {
	// Compile schema and panic on any compilation error
	if c, ok := r.validator.Validator.(schema.Compiler); ok {
		if err := c.Compile(); err != nil {
			return fmt.Errorf(": schema compilation error: %s", err)
		}
	}
	for field, subResource := range r.resources {
		if err := subResource.resource.Compile(); err != nil {
			if err.Error()[0] == ':' {
				// Check if I'm the direct ancestor of the raised sub-error
				return fmt.Errorf("%s%s", field, err)
			}
			return fmt.Errorf("%s.%s", field, err)
		}
	}
	return nil
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
	assertNotBound(name, r.resources, r.aliases)
	if f := s.validator.Validator.GetField(field); f == nil {
		log.Panicf("Cannot bind `%s' as sub-resource: field `%s' does not exist in the sub-resource'", name, field)
	}
	r.resources[name] = &subResource{
		field:    field,
		resource: s,
	}
	r.validator.schema[name] = schema.Field{
		ReadOnly: true,
		Validator: connection{
			path: "." + name,
		},
		Params: &schema.Params{
			Validators: map[string]schema.FieldValidator{
				"page": schema.Integer{
					Boundaries: &schema.Boundaries{Min: 1, Max: 1000},
				},
				"limit": schema.Integer{
					Boundaries: &schema.Boundaries{Min: 0, Max: 1000},
				},
				// TODO
				// "sort":   schema.String{}, // TODO proper sort validation
				// "filter": schema.String{}, // TODO proper filter validation
			},
		},
	}
	return s
}

// Alias adds an pre-built resource query on /<resource>/<alias>.
//
//     // Add a friendly alias to public posts on /users/:user_id/posts/public
//     // (equivalent to /users/:user_id/posts?filter={"public":true})
//     posts.Alias("public", url.Values{"where": []string{"{\"public\":true}"}})
//
// This method will panic an alias or a resource with the same name is already bound
func (r *Resource) Alias(name string, v url.Values) {
	assertNotBound(name, r.resources, r.aliases)
	r.aliases[name] = v
}

// GetAlias returns the alias set for the name if any
func (r *Resource) GetAlias(name string) (url.Values, bool) {
	a, found := r.aliases[name]
	return a, found
}

// Validator returns the resource's validator
func (r *Resource) Validator() schema.Validator {
	return r.validator
}

// Conf returns the resource's configuration
func (r *Resource) Conf() Conf {
	return r.conf
}

// Find implements Storer interface
func (r *Resource) Find(ctx context.Context, lookup *Lookup, page, perPage int) (*ItemList, error) {
	if r.storage == nil {
		return nil, ErrNoStorage
	}
	r.lockResources.Lock()
	defer r.lockResources.Unlock()
	return r.storage.Find(ctx, lookup, page, perPage)
}

// Insert implements Storer interface
func (r *Resource) Insert(ctx context.Context, items []*Item) error {
	if r.storage == nil {
		return ErrNoStorage
	}
	return r.storage.Insert(ctx, items)
}

// Update implements Storer interface
func (r *Resource) Update(ctx context.Context, item *Item, original *Item) error {
	if r.storage == nil {
		return ErrNoStorage
	}
	return r.storage.Update(ctx, item, original)
}

// Delete implements Storer interface
func (r *Resource) Delete(ctx context.Context, item *Item) error {
	if r.storage == nil {
		return ErrNoStorage
	}
	return r.storage.Delete(ctx, item)
}

// Clear implements Storer interface
func (r *Resource) Clear(ctx context.Context, lookup *Lookup) (int, error) {
	if r.storage == nil {
		return 0, ErrNoStorage
	}
	return r.storage.Clear(ctx, lookup)
}
