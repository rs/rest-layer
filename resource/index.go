package resource

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/rest-layer/schema"
)

// Index is an interface defining a type able to bind and retrieve resources
// from a resource graph.
type Index interface {
	// Bind a new resource at the "name" endpoint
	Bind(name string, s schema.Schema, h Storer, c Conf) *Resource
	// GetResource retrieves a given resource by it's path. For instance if a
	// resource "user" has a sub-resource "posts", a "users.posts" path can be
	// use to retrieve the posts resource.
	//
	// If a parent is given and the path starts with a dot, the lookup is
	// started at the parent's location instead of root's.
	GetResource(path string, parent *Resource) (*Resource, bool)
	// GetResources returns first level resources.
	GetResources() []*Resource
}

// Compiler is an optional interface for Index that's task is to prepare the
// index for usage. When the method exists, it's automatically called by
// rest.NewHandler(). When the resource package is used without the rest
// package, it's the user's responsibilty to call this method.
type Compiler interface {
	Compile() error
}

// index is the root of the resource graph.
type index struct {
	resources subResources
}

// NewIndex creates a new resource index.
func NewIndex() Index {
	return &index{
		resources: subResources{},
	}
}

// Bind a resource at the specified endpoint name.
func (i *index) Bind(name string, s schema.Schema, h Storer, c Conf) *Resource {
	assertNotBound(name, i.resources, nil)
	sr := newResource(name, s, h, c)
	i.resources.add(sr)
	return sr
}

// Compile the resource graph and report any error.
func (i *index) Compile() error {
	for _, r := range i.resources {
		if err := r.Compile(refChecker{i}); err != nil {
			sep := "."
			if err.Error()[0] == ':' {
				sep = ""
			}
			return fmt.Errorf("%s%s%s", r.name, sep, err)
		}
	}
	return nil
}

// GetResource retrieves a given resource by it's path. For instance if a resource "user" has a sub-resource "posts", a
// "users.posts" path can be use to retrieve the posts resource.
//
// If a parent is given and the path starts with a dot, the lookup is started at the
// parent's location instead of root's.
func (i *index) GetResource(path string, parent *Resource) (*Resource, bool) {
	resources := i.resources
	if len(path) > 0 && path[0] == '.' {
		if parent == nil {
			// If field starts with a dot and no parent is given, fail the lookup.
			return nil, false
		}
		path = path[1:]
		resources = parent.resources
	}
	var sr *Resource
	if strings.IndexByte(path, '.') == -1 {
		if sr = resources.get(path); sr == nil {
			return nil, false
		}
	} else {
		for _, comp := range strings.Split(path, ".") {
			if sr = resources.get(comp); sr == nil {
				return nil, false
			}
			resources = sr.resources
		}
	}
	return sr, true
}

// GetResources returns first level resources.
func (i *index) GetResources() []*Resource {
	return i.resources
}

// resourceLookup provides a wrapper for Index that implements the  schema.ReferenceChecker interface.
type refChecker struct {
	index Index
}

// ReferenceChecker implements the schema.ReferenceChecker interface.
func (rc refChecker) ReferenceChecker(path string) (schema.FieldValidator, schema.Validator) {
	rsc, exists := rc.index.GetResource(path, nil)
	if !exists {
		return nil, nil
	}
	validator := rsc.Schema().Fields["id"].Validator

	return schema.FieldValidatorFunc(func(value interface{}) (interface{}, error) {
		var id interface{}
		var err error

		if validator != nil {
			id, err = validator.Validate(value)
			if err != nil {
				return nil, err
			}
		} else {
			id = value
		}

		_, err = rsc.Get(context.TODO(), id)
		if err != nil {
			return nil, err
		}
		return id, nil
	}), rsc.Validator()
}

// assertNotBound asserts a given resource name is not already bound.
func assertNotBound(name string, resources subResources, aliases map[string]url.Values) {
	for _, r := range resources {
		if r.name == name {
			logPanicf(context.Background(), "Cannot bind `%s': already bound as resource'", name)
		}
	}
	if _, found := aliases[name]; found {
		logPanicf(context.Background(), "Cannot bind `%s': already bound as alias'", name)
	}
}
