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
		Type:        t.getObjectType(idx, r.Name(), r.Schema()),
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

func (t types) getListQuery(idx resource.Index, r *resource.Resource) *graphql.Field {
	return &graphql.Field{
		Description: fmt.Sprintf("Get a list of %s", r.Name()),
		Type:        graphql.NewList(t.getObjectType(idx, r.Name(), r.Schema())),
		Args: graphql.FieldConfigArgument{
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
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			page := 1
			// Default value on non HEAD request for perPage is -1 (pagination disabled)
			perPage := -1
			if l := r.Conf().PaginationDefaultLimit; l > 0 {
				perPage = l
			}
			if p, ok := p.Args["page"].(string); ok && p != "" {
				i, err := strconv.ParseUint(p, 10, 32)
				if err != nil {
					return nil, errors.New("invalid `limit` parameter")
				}
				page = int(i)
			}
			if l, ok := p.Args["limit"].(string); ok && l != "" {
				i, err := strconv.ParseUint(l, 10, 32)
				if err != nil {
					return nil, errors.New("invalid `limit` parameter")
				}
				perPage = int(i)
			}
			if perPage == -1 && page != 1 {
				return nil, errors.New("cannot use `page' parameter with no `limit' paramter on a resource with no default pagination size")
			}
			l := resource.NewLookup()
			if sort, ok := p.Args["sort"].(string); ok && sort != "" {
				if err := l.SetSort(sort, r.Validator()); err != nil {
					return nil, fmt.Errorf("invalid `sort` parameter: %v", err)
				}
			}
			if filter, ok := p.Args["filter"].(string); ok && filter != "" {
				if err := l.AddFilter(filter, r.Validator()); err != nil {
					return nil, fmt.Errorf("invalid `filter` parameter: %v", err)
				}
			}
			list, err := r.Find(p.Context, l, page, perPage)
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

func (t types) getSubQuery(idx resource.Index, r *resource.Resource, parentField string) *graphql.Field {
	return &graphql.Field{
		Type: t.getObjectType(idx, r.Name(), r.Schema()),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			parent, ok := p.Source.(map[string]interface{})
			if !ok {
				return nil, nil
			}
			item, err := r.Get(p.Context, parent[parentField])
			if err != nil {
				return nil, err
			}
			return item.Payload, nil
		},
	}
}
