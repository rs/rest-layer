package graphql

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/rs/rest-layer/resource"
)

func newRootQuery(idx resource.Index) *graphql.Object {
	t := types{}
	if c, ok := idx.(resource.Compiler); ok {
		if err := c.Compile(); err != nil {
			log.Fatal(err)
		}
	}
	flds := graphql.Fields{}
	for _, r := range idx.GetResources() {
		if r.Conf().IsModeAllowed(resource.Read) {
			flds[r.Name()] = t.getGetQuery(idx, r)
		}
		if r.Conf().IsModeAllowed(resource.List) {
			flds[r.Name()+"List"] = t.getListQuery(idx, r, nil)
			for _, a := range r.GetAliases() {
				params, _ := r.GetAlias(a)
				flds[r.Name()+strings.Title(a)] = t.getListQuery(idx, r, params)
			}
		}
	}
	return graphql.NewObject(graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: flds,
	})
}

func (t types) getGetQuery(idx resource.Index, r *resource.Resource) *graphql.Field {
	return &graphql.Field{
		Description: fmt.Sprintf("Get %s by id", r.Name()),
		Type:        t.getObjectType(idx, r),
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, ok := p.Args["id"].(string)
			if !ok {
				return nil, nil
			}
			item, err := r.Get(p.Context, id)
			if err != nil {
				return nil, err
			}
			return item.Payload, nil
		},
	}
}

var listArgs = graphql.FieldConfigArgument{
	"skip": &graphql.ArgumentConfig{
		Type: graphql.Int,
	},
	"page": &graphql.ArgumentConfig{
		Type: graphql.Int,
	},
	"limit": &graphql.ArgumentConfig{
		Type: graphql.Int,
	},
	"filter": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"sort": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

func listParamResolver(r *resource.Resource, p graphql.ResolveParams, params url.Values) (lookup *resource.Lookup, offset int, limit int, err error) {
	skip := 0
	page := 1
	// Default value on non HEAD request for limit is -1 (pagination disabled).
	limit = -1

	if l := r.Conf().PaginationDefaultLimit; l > 0 {
		limit = l
	}
	if i, ok := p.Args["skip"].(int); ok && i >= 0 {
		skip = i
	}
	if i, ok := p.Args["page"].(int); ok && i > 0 && i < 1000 {
		page = i
	}
	if i, ok := p.Args["limit"].(int); ok && i >= 0 && i < 1000 {
		limit = i
	}
	if page != 1 && limit == -1 {
		return nil, 0, 0, errors.New("cannot use `page' parameter with no `limit' parameter on a resource with no default pagination size")
	}
	offset = (page-1)*limit + skip
	lookup = resource.NewLookup()
	if sort, ok := p.Args["sort"].(string); ok && sort != "" {
		if err := lookup.SetSort(sort, r.Validator()); err != nil {
			return nil, 0, 0, fmt.Errorf("invalid `sort` parameter: %v", err)
		}
	}
	if filter, ok := p.Args["filter"].(string); ok && filter != "" {
		if err := lookup.AddFilter(filter, r.Validator()); err != nil {
			return nil, 0, 0, fmt.Errorf("invalid `filter` parameter: %v", err)
		}
	}
	if params != nil {
		if filter := params.Get("filter"); filter != "" {
			if err := lookup.AddFilter(filter, r.Validator()); err != nil {
				return nil, 0, 0, fmt.Errorf("invalid `filter` parameter: %v", err)
			}
		}
	}
	return
}

func (t types) getListQuery(idx resource.Index, r *resource.Resource, params url.Values) *graphql.Field {
	return &graphql.Field{
		Description: fmt.Sprintf("Get a list of %s", r.Name()),
		Type:        graphql.NewList(t.getObjectType(idx, r)),
		Args:        listArgs,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			lookup, offset, limit, err := listParamResolver(r, p, params)
			if err != nil {
				return nil, err
			}
			list, err := r.Find(p.Context, lookup, offset, limit)
			if err != nil {
				return nil, err
			}
			result := make([]map[string]interface{}, len(list.Items))
			for i, item := range list.Items {
				result[i] = item.Payload
			}
			return result, nil
		},
	}
}
