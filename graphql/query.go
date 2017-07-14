package graphql

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema/query"
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

func listParamResolver(r *resource.Resource, p graphql.ResolveParams, params url.Values) (q *query.Query, err error) {
	skip := 0
	page := 1
	// Default value on non HEAD request for limit is -1 (pagination disabled).
	limit := -1
	if i, ok := p.Args["skip"].(int); ok && i >= 0 {
		skip = i
	}
	if i, ok := p.Args["page"].(int); ok && i > 0 && i < 1000 {
		page = i
	}
	if i, ok := p.Args["limit"].(int); ok && i >= 0 && i < 1000 {
		limit = i
	} else if l := r.Conf().PaginationDefaultLimit; l > 0 {
		limit = l
	}
	if page != 1 && limit == -1 {
		return nil, errors.New("cannot use `page' parameter with no `limit' parameter on a resource with no default pagination size")
	}
	q = &query.Query{}
	q.Window = query.Page(page, limit, skip)
	if sort, ok := p.Args["sort"].(string); ok && sort != "" {
		s, err := query.ParseSort(sort)
		if err == nil {
			err = s.Validate(r.Validator())
		}
		if err != nil {
			return nil, fmt.Errorf("invalid `sort` parameter: %v", err)
		}
		q.Sort = s
	}
	if filter, ok := p.Args["filter"].(string); ok && filter != "" {
		p, err := query.ParsePredicate(filter)
		if err == nil {
			err = p.Validate(r.Validator())
		}
		if err != nil {
			return nil, fmt.Errorf("invalid `filter` parameter: %v", err)
		}
		q.Predicate = p
	}
	if params != nil {
		if filter := params.Get("filter"); filter != "" {
			p, err := query.ParsePredicate(filter)
			if err == nil {
				err = p.Validate(r.Validator())
			}
			if err != nil {
				return nil, fmt.Errorf("invalid `filter` parameter: %v", err)
			}
			if len(p) > 0 {
				q.Predicate = append(q.Predicate, p...)
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
			q, err := listParamResolver(r, p, params)
			if err != nil {
				return nil, err
			}
			list, err := r.Find(p.Context, q)
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
