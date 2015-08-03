package rest

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

// RootResource is the root of the resource graph
type RootResource struct {
	resources map[string]*subResource
}

// New creates a new root resource
func New() *RootResource {
	return &RootResource{
		resources: map[string]*subResource{},
	}
}

// Bind a resource at the specified endpoint name
func (r *RootResource) Bind(name string, s *Resource) *Resource {
	assertNotBound(name, r.resources, nil)
	r.resources[name] = &subResource{resource: s}
	return s
}

// Compile the resource graph and report any error
func (r *RootResource) Compile() error {
	return compileResourceGraph(r.resources)
}

// GetResource retrives a given resource by it's path.
// For instance if a resource user has a sub-resource posts,
// a users.posts path can be use to retrieve the posts resource.
func (r *RootResource) GetResource(path string) *Resource {
	resources := r.resources
	var resource *Resource
	for _, comp := range strings.Split(path, ".") {
		if subResource, found := resources[comp]; found {
			resource = subResource.resource
			resources = resource.resources
		} else {
			return nil
		}
	}
	return resource
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
