package graphql

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/graphql-go/graphql"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
)

func newRootQuery(idx resource.Index) *graphql.Object {
	t := types{}
	if c, ok := idx.(schema.Compiler); ok {
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
			flds[r.Name()+"List"] = t.getListQuery(idx, r)
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

func listParamResolver(r *resource.Resource, p graphql.ResolveParams) (lookup *resource.Lookup, page int, perPage int, err error) {
	page = 1
	// Default value on non HEAD request for perPage is -1 (pagination disabled)
	perPage = -1
	if l := r.Conf().PaginationDefaultLimit; l > 0 {
		perPage = l
	}
	if p, ok := p.Args["page"].(string); ok && p != "" {
		i, err := strconv.ParseUint(p, 10, 32)
		if err != nil {
			return nil, 0, 0, errors.New("invalid `limit` parameter")
		}
		page = int(i)
	}
	if l, ok := p.Args["limit"].(string); ok && l != "" {
		i, err := strconv.ParseUint(l, 10, 32)
		if err != nil {
			return nil, 0, 0, errors.New("invalid `limit` parameter")
		}
		perPage = int(i)
	}
	if perPage == -1 && page != 1 {
		return nil, 0, 0, errors.New("cannot use `page' parameter with no `limit' paramter on a resource with no default pagination size")
	}
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
	return
}

func (t types) getListQuery(idx resource.Index, r *resource.Resource) *graphql.Field {
	return &graphql.Field{
		Description: fmt.Sprintf("Get a list of %s", r.Name()),
		Type:        graphql.NewList(t.getObjectType(idx, r)),
		Args:        listArgs,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			lookup, page, perPage, err := listParamResolver(r, p)
			if err != nil {
				return nil, err
			}
			list, err := r.Find(p.Context, lookup, page, perPage)
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
