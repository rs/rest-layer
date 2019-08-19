package resource

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
)

// Resource holds information about a class of items exposed on the API.
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

// get gets a sub resource by its name.
func (sr subResources) get(name string) *Resource {
	i := sort.Search(len(sr), func(i int) bool {
		return sr[i].name >= name
	})
	if i >= len(sr) {
		return nil
	}
	r := sr[i]
	if r.name != name {
		return nil
	}
	return r
}

// add adds the resource to the subResources in a pre-sorted way.
func (sr *subResources) add(rsrc *Resource) {
	for i, r := range *sr {
		if rsrc.name < r.name {
			*sr = append((*sr)[:i], append(subResources{rsrc}, (*sr)[i:]...)...)
			return
		}
	}
	*sr = append(*sr, rsrc)
}

// validatorFallback wraps a validator and fallback on given schema if the GetField
// returns nil on a given name.
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

// newResource creates a new resource with provided spec, handler and config.
func newResource(name string, s schema.Schema, h Storer, c Conf) *Resource {
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
// intermediate resources separated by dots (i.e.: res1.res2.res3).
func (r *Resource) Path() string {
	return r.path
}

// ParentField returns the name of the field on which the resource is bound to
// its parent if any.
func (r *Resource) ParentField() string {
	return r.parentField
}

// Compile the resource graph and report any error.
func (r *Resource) Compile(rc schema.ReferenceChecker) error {
	// Compile schema and panic on any compilation error.
	if c, ok := r.validator.Validator.(schema.Compiler); ok {
		if err := c.Compile(rc); err != nil {
			return fmt.Errorf(": schema compilation error: %s", err)
		}
	}
	for _, r := range r.resources {
		if err := r.Compile(rc); err != nil {
			if err.Error()[0] == ':' {
				// Check if I'm the direct ancestor of the raised sub-error.
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
		logPanicf(nil, "Cannot bind `%s' as sub-resource: field `%s' does not exist in the sub-resource'", name, field)
	}
	sr := newResource(name, s, h, c)
	sr.parentField = field
	sr.path = r.path + "." + name
	r.resources.add(sr)
	r.validator.fallback.Fields[name] = schema.Field{
		ReadOnly: true,
		Validator: &schema.Connection{
			Path:      "." + name,
			Field:     field,
			Validator: sr.validator,
		},
		Params: schema.Params{
			"skip": schema.Param{
				Description: "The number of items to skip",
				Validator: schema.Integer{
					Boundaries: &schema.Boundaries{Min: 0},
				},
			},
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

// GetResources returns first level resources.
func (r *Resource) GetResources() []*Resource {
	return r.resources
}

// Alias adds an pre-built resource query on /<resource>/<alias>.
//
//     // Add a friendly alias to public posts on /users/:user_id/posts/public
//     // (equivalent to /users/:user_id/posts?filter={"public":true})
//     posts.Alias("public", url.Values{"where": []string{"{\"public\":true}"}})
//
// This method will panic an alias or a resource with the same name is already bound.
func (r *Resource) Alias(name string, v url.Values) {
	assertNotBound(name, r.resources, r.aliases)
	r.aliases[name] = v
}

// GetAlias returns the alias set for the name if any.
func (r *Resource) GetAlias(name string) (url.Values, bool) {
	a, found := r.aliases[name]
	return a, found
}

// GetAliases returns all the alias names set on the resource.
func (r *Resource) GetAliases() []string {
	n := make([]string, 0, len(r.aliases))
	for a := range r.aliases {
		n = append(n, a)
	}
	return n
}

// Schema returns the resource's schema.
func (r *Resource) Schema() schema.Schema {
	return r.schema
}

// Validator returns the resource's validator.
func (r *Resource) Validator() schema.Validator {
	return r.validator
}

// Conf returns the resource's configuration.
func (r *Resource) Conf() Conf {
	return r.conf
}

// Use attaches an event handler to the resource. This event handler must
// implement on of the resource.*EventHandler interface or this method returns
// an error.
func (r *Resource) Use(e interface{}) error {
	return r.hooks.use(e)
}

// Get get one item by its id. If item is not found, ErrNotFound error is
// returned.
func (r *Resource) Get(ctx context.Context, id interface{}) (item *Item, err error) {
	if LoggerLevel <= LogLevelDebug && Logger != nil {
		defer func(t time.Time) {
			Logger(ctx, LogLevelDebug, fmt.Sprintf("%s.Get(%v)", r.path, id), map[string]interface{}{
				"duration": time.Since(t),
				"error":    err,
			})
		}(time.Now())
	}
	if err = r.hooks.onGet(ctx, id); err == nil {
		item, err = r.storage.Get(ctx, id)
	}
	r.hooks.onGot(ctx, &item, &err)
	return
}

// MultiGet get some items by their id and return them in the same order. If one
// or more item(s) is not found, their slot in the response is set to nil.
func (r *Resource) MultiGet(ctx context.Context, ids []interface{}) (items []*Item, err error) {
	if LoggerLevel <= LogLevelDebug && Logger != nil {
		defer func(t time.Time) {
			Logger(ctx, LogLevelDebug, fmt.Sprintf("%s.MultiGet(%v)", r.path, ids), map[string]interface{}{
				"duration": time.Since(t),
				"found":    len(items),
				"error":    err,
			})
		}(time.Now())
	}
	errs := make([]error, len(ids))
	for i, id := range ids {
		errs[i] = r.hooks.onGet(ctx, id)
		if err == nil && errs[i] != nil {
			// first pre-hook error is the global error.
			err = errs[i]
		}
	}
	// Perform the storage request if none of the pre-hook returned an err.
	if err == nil {
		items, err = r.storage.MultiGet(ctx, ids)
	}
	var errOverwrite error
	for i := range ids {
		var _item *Item
		if len(items) > i {
			_item = items[i]
		}
		// Give the pre-hook error for this id or global otherwise.
		_err := errs[i]
		if _err == nil {
			_err = err
		}
		r.hooks.onGot(ctx, &_item, &_err)
		if errOverwrite == nil && _err != errs[i] {
			errOverwrite = _err // apply change done on the first error.
		}
		if _err == nil && len(items) > i && _item != items[i] {
			items[i] = _item // apply changes done by hooks if any.
		}
	}
	if errOverwrite != nil {
		err = errOverwrite
	}
	if err != nil {
		items = nil
	}
	return
}

// Find calls the Find method on the storage handler with the corresponding pre/post hooks.
func (r *Resource) Find(ctx context.Context, q *query.Query) (list *ItemList, err error) {
	return r.find(ctx, q, false)
}

// FindWithTotal calls the Find method on the storage handler with the
// corresponding pre/post hooks. If the storage is not able to compute the
// total, this method will call the Count method on the storage. If the storage
// Find does not compute the total and the Counter interface is not implemented,
// an ErrNotImplemented error is returned.
func (r *Resource) FindWithTotal(ctx context.Context, q *query.Query) (list *ItemList, err error) {
	return r.find(ctx, q, true)
}

func (r *Resource) find(ctx context.Context, q *query.Query, forceTotal bool) (list *ItemList, err error) {
	if LoggerLevel <= LogLevelDebug && Logger != nil {
		defer func(t time.Time) {
			found := -1
			if list != nil {
				found = len(list.Items)
			}
			Logger(ctx, LogLevelDebug, fmt.Sprintf("%s.Find(...)", r.path), map[string]interface{}{
				"duration": time.Since(t),
				"found":    found,
				"error":    err,
			})
		}(time.Now())
	}
	if err = r.hooks.onFind(ctx, q); err == nil {
		list, err = r.storage.Find(ctx, q)
		if err == nil && list.Total == -1 && forceTotal {
			// Send a query with no window so the storage won't be tempted to
			// count within the window.
			list.Total, err = r.storage.Count(ctx, &query.Query{Predicate: q.Predicate})
		}
	}
	r.hooks.onFound(ctx, q, &list, &err)
	return
}

// Insert implements Storer interface.
func (r *Resource) Insert(ctx context.Context, items []*Item) (err error) {
	if LoggerLevel <= LogLevelDebug && Logger != nil {
		defer func(t time.Time) {
			Logger(ctx, LogLevelDebug, fmt.Sprintf("%s.Insert(items[%d])", r.path, len(items)), map[string]interface{}{
				"duration": time.Since(t),
				"error":    err,
			})
		}(time.Now())
	}
	if err = r.hooks.onInsert(ctx, items); err == nil {
		if err = recalcEtag(items); err == nil {
			err = r.storage.Insert(ctx, items)
		}
	}
	r.hooks.onInserted(ctx, items, &err)
	return
}

func recalcEtag(items []*Item) error {
	if items == nil {
		return nil
	}

	for _, v := range items {
		if v == nil {
			continue
		}
		etag, err := genEtag(v.Payload)
		if err != nil {
			return err
		}
		v.ETag = etag
	}
	return nil
}

// Update implements Storer interface.
func (r *Resource) Update(ctx context.Context, item *Item, original *Item) (err error) {
	if LoggerLevel <= LogLevelDebug && Logger != nil {
		defer func(t time.Time) {
			Logger(ctx, LogLevelDebug, fmt.Sprintf("%s.Update(%v, %v)", r.path, item.ID, original.ID), map[string]interface{}{
				"duration": time.Since(t),
				"error":    err,
			})
		}(time.Now())
	}
	if err = r.hooks.onUpdate(ctx, item, original); err == nil {
		if err = recalcEtag([]*Item{item}); err == nil {
			err = r.storage.Update(ctx, item, original)
		}
	}
	r.hooks.onUpdated(ctx, item, original, &err)
	return
}

// Delete implements Storer interface.
func (r *Resource) Delete(ctx context.Context, item *Item) (err error) {
	if LoggerLevel <= LogLevelDebug && Logger != nil {
		defer func(t time.Time) {
			Logger(ctx, LogLevelDebug, fmt.Sprintf("%s.Delete(%v)", r.path, item.ID), map[string]interface{}{
				"duration": time.Since(t),
				"error":    err,
			})
		}(time.Now())
	}
	if err = r.hooks.onDelete(ctx, item); err == nil {
		err = r.storage.Delete(ctx, item)
	}
	r.hooks.onDeleted(ctx, item, &err)
	return
}

// Clear implements Storer interface.
func (r *Resource) Clear(ctx context.Context, q *query.Query) (deleted int, err error) {
	if LoggerLevel <= LogLevelDebug && Logger != nil {
		defer func(t time.Time) {
			Logger(ctx, LogLevelDebug, fmt.Sprintf("%s.Clear(%v)", r.path, q), map[string]interface{}{
				"duration": time.Since(t),
				"deleted":  deleted,
				"error":    err,
			})
		}(time.Now())
	}
	if err = r.hooks.onClear(ctx, q); err == nil {
		deleted, err = r.storage.Clear(ctx, q)
	}
	r.hooks.onCleared(ctx, q, &deleted, &err)
	return
}
