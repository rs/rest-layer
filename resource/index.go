package resource

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

// Index is an interface defining a type able to bind and retrieve resources
// from a resource graph.
type Index interface {
	// Bind a new resource at the "name" endpoint
	Bind(name string, s *Resource) *Resource
	// GetResource retrives a given resource and its parent field identifier by it's path.
	// For instance if a resource user has a sub-resource posts,
	// a users.posts path can be use to retrieve the posts resource.
	GetResource(path string, parent *Resource) (*Resource, string, bool)
}

// index is the root of the resource graph
type index struct {
	resources map[string]*subResource
}

// NewIndex creates a new resource index
func NewIndex() Index {
	return &index{
		resources: map[string]*subResource{},
	}
}

// Bind a resource at the specified endpoint name
func (r *index) Bind(name string, s *Resource) *Resource {
	assertNotBound(name, r.resources, nil)
	r.resources[name] = &subResource{resource: s}
	return s
}

// Compile the resource graph and report any error
func (r *index) Compile() error {
	return compileResourceGraph(r.resources)
}

// GetResource retrives a given resource and its parent field identifier by it's path.
// For instance if a resource user has a sub-resource posts,
// a users.posts path can be use to retrieve the posts resource.
//
// If a parent is given and the path starts with a dot, the lookup is started at the
// parent's location instead of root's.
func (r *index) GetResource(path string, parent *Resource) (*Resource, string, bool) {
	resources := r.resources
	field := ""
	if len(path) > 0 && path[0] == '.' {
		if parent == nil {
			// If field starts with a dot and no parent is given, fail the lookup
			return nil, "", false
		}
		path = path[1:]
		resources = parent.resources
	}
	var resource *Resource
	for _, comp := range strings.Split(path, ".") {
		if subResource, found := resources[comp]; found {
			resource = subResource.resource
			field = subResource.field
			resources = resource.resources
		} else {
			return nil, "", false
		}
	}
	return resource, field, true
}

func compileResourceGraph(resources map[string]*subResource) error {
	for field, subResource := range resources {
		if err := subResource.resource.Compile(); err != nil {
			sep := "."
			if err.Error()[0] == ':' {
				sep = ""
			}
			return fmt.Errorf("%s%s%s", field, sep, err)
		}
	}
	return nil
}

// assertNotBound asserts a given resource name is not already bound
func assertNotBound(name string, resources map[string]*subResource, aliases map[string]url.Values) {
	if _, found := resources[name]; found {
		log.Panicf("Cannot bind `%s': already bound as resource'", name)
	}
	if _, found := aliases[name]; found {
		log.Panicf("Cannot bind `%s': already bound as alias'", name)
	}
}
