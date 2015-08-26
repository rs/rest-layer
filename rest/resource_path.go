package rest

import (
	"strings"
	"sync"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"golang.org/x/net/context"
)

// ResourcePath is the list of ResourcePathComponent leading to the requested resource
type ResourcePath []*ResourcePathComponent

// ResourcePathComponent represents the path of resource and sub-resources of a given request's resource
type ResourcePathComponent struct {
	// Name is the endpoint name used to bind the resource
	Name string
	// Field is the resource's field used to filter targetted resource
	Field string
	// Value holds the resource's id value
	Value interface{}
	// Resource references the resource
	Resource *resource.Resource
}

var resourcePathComponentPool = sync.Pool{
	New: func() interface{} {
		return &ResourcePathComponent{}
	},
}

// Prepend add the given resource using the provided field and value as a "ghost" resource
// prefix to the resource path.
//
// The effect will be a 404 error if the doesn't have an item with the id matching to the
// provided value.
//
// This will also require that all subsequent resources in the path have this resource's
// "value" set on their "field" field.
//
// Finaly, all created resources at this path will also have this field and value set by default.
func (p *ResourcePath) Prepend(rsrc *resource.Resource, field string, value interface{}) {
	rp := resourcePathComponentPool.Get().(*ResourcePathComponent)
	rp.Field = field
	rp.Value = value
	rp.Resource = rsrc
	// Prepent the resource path with the user resource
	*p = append(ResourcePath{rp}, *p...)
}

func (p *ResourcePath) append(rsrc *resource.Resource, field string, value interface{}, name string) {
	rp := resourcePathComponentPool.Get().(*ResourcePathComponent)
	rp.Name = name
	rp.Field = field
	rp.Value = value
	rp.Resource = rsrc
	*p = append(*p, rp)
}

func (p *ResourcePath) clear() {
	for _, rp := range *p {
		rp.Name = ""
		rp.Field = ""
		rp.Value = nil
		rp.Resource = nil
		resourcePathComponentPool.Put(rp)
	}
	*p = nil
}

// ParentsExist checks if the each intermediate parents in the path exist and
// return either a ErrNotFound or an error returned by on of the intermediate
// resource.
func (p ResourcePath) ParentsExist(ctx context.Context) error {
	// First we check that we have no field conflict on the path (i.e.: two path
	// components defining the same field with a different value)
	fields := map[string]interface{}{}
	for _, rp := range p {
		if val, found := fields[rp.Field]; found && val != rp.Value {
			return &Error{404, "Resource Path Conflict", nil}
		}
		fields[rp.Field] = rp.Value
	}

	// Check parents existence
	parents := len(p) - 1
	if parents <= 0 {
		return nil
	}
	q := schema.Query{}
	c := make(chan error)
	for _, rp := range p[:parents] {
		if rp.Value == nil {
			continue
		}
		// Create a lookup with the parent path fields + the current path id
		l := resource.NewLookup()
		lq := append(q[:], schema.Equal{Field: "id", Value: rp.Value})
		l.AddQuery(lq)
		// Execute all intermediate checkes in concurence
		go func() {
			// Check if the resource exists
			list, err := rp.Resource.Find(ctx, l, 1, 1)
			if err != nil {
				c <- err
			} else if len(list.Items) == 0 {
				c <- &Error{404, "Parent Resource Not Found", nil}
			} else {
				c <- nil
			}
		}()
		// Push the resource field=value for the next hops
		q = append(q, schema.Equal{Field: rp.Field, Value: rp.Value})
	}
	// Fail on first error
	for i := 0; i < parents; i++ {
		if err := <-c; err != nil {
			return err
		}
	}
	return nil
}

// Path returns the path to the resource to be used with resource.Root.GetResource
func (p ResourcePath) Path() string {
	path := []string{}
	for _, c := range p {
		if c.Name != "" {
			path = append(path, c.Name)
		}
	}
	return strings.Join(path, ".")
}

// Values returns all the key=value pairs defined by the resource path
func (p ResourcePath) Values() map[string]interface{} {
	v := map[string]interface{}{}
	for _, rp := range p {
		if _, found := v[rp.Field]; !found && rp.Value != nil {
			v[rp.Field] = rp.Value
		}
	}
	return v
}
