package resource

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/rs/rest-layer/schema"
)

// Index is an interface defining a type able to bind and retrieve resources
// from a resource graph.
type Index interface {
	// Bind a new resource at the "name" endpoint
	Bind(name string, v schema.Validator, h Storer, c Conf) *Resource
	// GetResource retrives a given resource and its parent field identifier by it's path.
	// For instance if a resource user has a sub-resource posts,
	// a users.posts path can be use to retrieve the posts resource.
	GetResource(path string, parent *Resource) (*Resource, string, bool)
}

// index is the root of the resource graph
type index struct {
	resources subResources
}

// NewIndex creates a new resource index
func NewIndex() Index {
	return &index{
		resources: subResources{},
	}
}

// Bind a resource at the specified endpoint name
func (r *index) Bind(name string, v schema.Validator, h Storer, c Conf) *Resource {
	assertNotBound(name, r.resources, nil)
	s := new(name, v, h, c)
	r.resources = append(r.resources, &subResource{resource: s})
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
		if subResource := resources.get(comp); subResource != nil {
			resource = subResource.resource
			field = subResource.field
			resources = resource.resources
		} else {
			return nil, "", false
		}
	}
	return resource, field, true
}

func compileResourceGraph(resources subResources) error {
	for _, subResource := range resources {
		if err := subResource.resource.Compile(); err != nil {
			sep := "."
			if err.Error()[0] == ':' {
				sep = ""
			}
			return fmt.Errorf("%s%s%s", subResource.resource.name, sep, err)
		}
	}
	return nil
}

// assertNotBound asserts a given resource name is not already bound
func assertNotBound(name string, resources []*subResource, aliases map[string]url.Values) {
	for _, r := range resources {
		if r.resource.name == name {
			log.Panicf("Cannot bind `%s': already bound as resource'", name)
		}
	}
	if _, found := aliases[name]; found {
		log.Panicf("Cannot bind `%s': already bound as alias'", name)
	}
}
