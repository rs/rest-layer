package resource

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/xlog"
	"golang.org/x/net/context"
)

// Resource holds information about a class of items exposed on the API
type Resource struct {
	parentField string
	name        string
	path        string
	schema      schema.Schema
	validator   validatorFallback
	storage     storageHandler
	conf        Conf
	resources   subResources
	aliases     map[string]url.Values
	hooks       eventHandler
}

type subResources []*Resource

func (sr subResources) get(name string) *Resource {
	// TODO pre-sort and use sort package to search
	for _, r := range sr {
		if r.name == name {
			return r
		}
	}
	return nil
}

// validatorFallback wraps a validator and fallback on given schema if the GetField
// returns nil on a given name
type validatorFallback struct {
	schema.Validator
	fallback schema.Schema
}

func (v validatorFallback) GetField(name string) *schema.Field {
	if f := v.Validator.GetField(name); f != nil {
		return f
	}
	return v.fallback.GetField(name)
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

// new creates a new resource with provided spec, handler and config
func new(name string, s schema.Schema, h Storer, c Conf) *Resource {
	return &Resource{
		name:   name,
		path:   name,
		schema: s,
		validator: validatorFallback{
			Validator: s,
			fallback:  schema.Schema{Fields: schema.Fields{}},
		},
		storage:   storageWrapper{h},
		conf:      c,
		resources: subResources{},
		aliases:   map[string]url.Values{},
	}
}

// Name returns the name of the resource
func (r *Resource) Name() string {
	return r.name
}

// Path returns the full path of the resource composed of names of each
// intermediate resources separated by dots (i.e.: res1.res2.res3)
func (r *Resource) Path() string {
	return r.path
}

// ParentField returns the name of the field on which the resource is bound to its parent if any.
func (r *Resource) ParentField() string {
	return r.parentField
}

// Compile the resource graph and report any error
func (r *Resource) Compile() error {
	// Compile schema and panic on any compilation error
	if c, ok := r.validator.Validator.(schema.Compiler); ok {
		if err := c.Compile(); err != nil {
			return fmt.Errorf(": schema compilation error: %s", err)
		}
	}
	for _, r := range r.resources {
		if err := r.Compile(); err != nil {
			if err.Error()[0] == ':' {
				// Check if I'm the direct ancestor of the raised sub-error
				return fmt.Errorf("%s%s", r.name, err)
			}
			return fmt.Errorf("%s.%s", r.name, err)
		}
	}
	return nil
}

// Bind a sub-resource with the provided name. The field parameter defines the parent
// resource's which contains the sub resource id.
//
//     users := api.Bind("users", userSchema, userHandler, userConf)
//     // Bind a sub resource on /users/:user_id/posts[/:post_id]
//     // and reference the user on each post using the "user" field.
//     posts := users.Bind("posts", "user", postSchema, postHandler, postConf)
//
// This method will panic an alias or a resource with the same name is already bound
// or if the specified field doesn't exist in the parent resource spec.
func (r *Resource) Bind(name, field string, s schema.Schema, h Storer, c Conf) *Resource {
	assertNotBound(name, r.resources, r.aliases)
	if f := s.GetField(field); f == nil {
		log.Panicf("Cannot bind `%s' as sub-resource: field `%s' does not exist in the sub-resource'", name, field)
	}
	sr := new(name, s, h, c)
	sr.parentField = field
	sr.path = r.path + "." + name
	r.resources = append(r.resources, sr)
	r.validator.fallback.Fields[name] = schema.Field{
		ReadOnly: true,
		Validator: connection{
			path: "." + name,
		},
		Params: schema.Params{
			"page": schema.Param{
				Description: "The page number",
				Validator: schema.Integer{
					Boundaries: &schema.Boundaries{Min: 1, Max: 1000},
				},
			},
			"limit": schema.Param{
				Description: "The number of items to return per page",
				Validator: schema.Integer{
					Boundaries: &schema.Boundaries{Min: 0, Max: 1000},
				},
			},
			"sort": schema.Param{
				Description: "The field(s) to sort on",
				Validator:   schema.String{},
			},
			"filter": schema.Param{
				Description: "The filter query",
				Validator:   schema.String{},
			},
		},
	}
	return sr
}

// GetResources returns first level resources
func (r *Resource) GetResources() []*Resource {
	return r.resources
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

// GetAliases returns all the alias names set on the resource
func (r *Resource) GetAliases() []string {
	n := make([]string, 0, len(r.aliases))
	for a := range r.aliases {
		n = append(n, a)
	}
	return n
}

// Schema returns the resource's schema
func (r *Resource) Schema() schema.Schema {
	return r.schema
}

// Validator returns the resource's validator
func (r *Resource) Validator() schema.Validator {
	return r.validator
}

// Conf returns the resource's configuration
func (r *Resource) Conf() Conf {
	return r.conf
}

// Use attaches an event handler to the resource. This event
// handler must implement on of the resource.*EventHandler interface
// or this method returns an error.
func (r *Resource) Use(e interface{}) error {
	return r.hooks.use(e)
}

// Get get one item by its id. If item is not found, ErrNotFound error is returned
func (r *Resource) Get(ctx context.Context, id interface{}) (item *Item, err error) {
	defer func(t time.Time) {
		xlog.FromContext(ctx).Debugf("%s.Get(%v)", r.path, id, xlog.F{
			"duration": time.Since(t),
			"error":    err,
		})
	}(time.Now())
	if err = r.hooks.onGet(ctx, id); err == nil {
		item, err = r.storage.Get(ctx, id)
	}
	r.hooks.onGot(ctx, &item, &err)
	return
}

// MultiGet get some items by their id and return them in the same order. If one or more item(s)
// is not found, their slot in the response is set to nil.
func (r *Resource) MultiGet(ctx context.Context, ids []interface{}) (items []*Item, err error) {
	defer func(t time.Time) {
		xlog.FromContext(ctx).Debugf("%s.MultiGet(%v)", r.path, ids, xlog.F{
			"duration": time.Since(t),
			"found":    len(items),
			"error":    err,
		})
	}(time.Now())
	for _, id := range ids {
		err = r.hooks.onGet(ctx, id)
		if err != nil {
			return
		}
	}
	if err == nil {
		items, err = r.storage.MultiGet(ctx, ids)
	}
	for i, item := range items {
		_item := item
		r.hooks.onGot(ctx, &_item, &err)
		if _item != item {
			items[i] = _item // apply changes done by hooks if any
		}
	}
	return
}

// Find implements Storer interface
func (r *Resource) Find(ctx context.Context, lookup *Lookup, page, perPage int) (list *ItemList, err error) {
	defer func(t time.Time) {
		found := -1
		if list != nil {
			found = len(list.Items)
		}
		xlog.FromContext(ctx).Debugf("%s.Find(..., %d, %d)", r.path, page, perPage, xlog.F{
			"duration": time.Since(t),
			"found":    found,
			"error":    err,
		})
	}(time.Now())
	if err = r.hooks.onFind(ctx, lookup, page, perPage); err == nil {
		list, err = r.storage.Find(ctx, lookup, page, perPage)
	}
	r.hooks.onFound(ctx, lookup, &list, &err)
	return
}

// Insert implements Storer interface
func (r *Resource) Insert(ctx context.Context, items []*Item) (err error) {
	defer func(t time.Time) {
		xlog.FromContext(ctx).Debugf("%s.Insert(items[%d])", r.path, len(items), xlog.F{
			"duration": time.Since(t),
			"error":    err,
		})
	}(time.Now())
	if err = r.hooks.onInsert(ctx, items); err == nil {
		err = r.storage.Insert(ctx, items)
	}
	r.hooks.onInserted(ctx, items, &err)
	return
}

// Update implements Storer interface
func (r *Resource) Update(ctx context.Context, item *Item, original *Item) (err error) {
	defer func(t time.Time) {
		xlog.FromContext(ctx).Debugf("%s.Update(%v, %v)", r.path, item.ID, original.ID, xlog.F{
			"duration": time.Since(t),
			"error":    err,
		})
	}(time.Now())
	if err = r.hooks.onUpdate(ctx, item, original); err == nil {
		err = r.storage.Update(ctx, item, original)
	}
	r.hooks.onUpdated(ctx, item, original, &err)
	return
}

// Delete implements Storer interface
func (r *Resource) Delete(ctx context.Context, item *Item) (err error) {
	defer func(t time.Time) {
		xlog.FromContext(ctx).Debugf("%s.Delete(%v)", r.path, item.ID, xlog.F{
			"duration": time.Since(t),
			"error":    err,
		})
	}(time.Now())
	if err = r.hooks.onDelete(ctx, item); err == nil {
		err = r.storage.Delete(ctx, item)
	}
	r.hooks.onDeleted(ctx, item, &err)
	return
}

// Clear implements Storer interface
func (r *Resource) Clear(ctx context.Context, lookup *Lookup) (deleted int, err error) {
	defer func(t time.Time) {
		xlog.FromContext(ctx).Debugf("%s.Clear(%v)", r.path, lookup, xlog.F{
			"duration": time.Since(t),
			"deleted":  deleted,
			"error":    err,
		})
	}(time.Now())
	if err = r.hooks.onClear(ctx, lookup); err == nil {
		deleted, err = r.storage.Clear(ctx, lookup)
	}
	r.hooks.onCleared(ctx, lookup, &deleted, &err)
	return
}
