package graphql

import (
	"fmt"
	"log"

	"github.com/graphql-go/graphql"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
)

type types map[string]*graphql.Object

// getObjectType returns a graphql object type definition from a REST layer
// schema.
func (t types) getObjectType(idx resource.Index, r *resource.Resource) *graphql.Object {
	// Memoize types by their name so we don't create several instance of the
	// same resource.
	name := r.Name()
	o := t[name]
	if o == nil {
		o = graphql.NewObject(graphql.ObjectConfig{
			Name:        name,
			Description: r.Schema().Description,
			Fields:      getFields(idx, r.Schema()),
		})
		t[name] = o
		t.addConnections(o, idx, r)
	}
	return o
}

// addConnections adds connections fields to the object afterward to prevent
// from dead loops.
func (t types) addConnections(o *graphql.Object, idx resource.Index, r *resource.Resource) {
	// Add sub field references.
	for name, def := range r.Schema().Fields {
		if ref, ok := def.Validator.(*schema.Reference); ok {
			sr, found := idx.GetResource(ref.Path, nil)
			if !found {
				log.Panicf("resource reference not found: %s", ref.Path)
			}
			o.AddFieldConfig(name, &graphql.Field{
				Description: def.Description,
				Type:        t.getObjectType(idx, sr),
				Args:        getFArgs(def.Params),
				Resolve:     getSubFieldResolver(name, sr, def),
			})
		}
	}
	// Add sub resources.
	for _, sr := range r.GetResources() {
		name := sr.Name()
		o.AddFieldConfig(name, &graphql.Field{
			Description: fmt.Sprintf("Connection to %s", name),
			Type:        graphql.NewList(t.getObjectType(idx, sr)),
			Args:        listArgs,
			Resolve:     getSubResourceResolver(sr),
		})
	}
}

func getSubFieldResolver(parentField string, r *resource.Resource, f schema.Field) graphql.FieldResolveFn {
	s, serialize := f.Validator.(schema.FieldSerializer)
	return func(p graphql.ResolveParams) (data interface{}, err error) {
		parent, ok := p.Source.(map[string]interface{})
		if !ok {
			return nil, nil
		}
		var item *resource.Item
		// Get sub field resource.
		item, err = r.Get(p.Context, parent[parentField])
		if err != nil {
			return nil, err
		}
		data = item.Payload
		if f.Handler != nil {
			data, err = f.Handler(p.Context, data, p.Args)
		}
		if err == nil && serialize {
			data, err = s.Serialize(data)
		}
		return data, err
	}
}

func getSubResourceResolver(r *resource.Resource) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		parent, ok := p.Source.(map[string]interface{})
		if !ok {
			return nil, nil
		}
		q, err := listParamResolver(r, p, nil)
		if err != nil {
			return nil, err
		}
		// Limit the connection to parent's owned.
		q.Predicate = append(q.Predicate, &query.Equal{Field: r.ParentField(), Value: parent["id"]})
		list, err := r.Find(p.Context, q)
		if err != nil {
			return nil, err
		}
		result := make([]map[string]interface{}, len(list.Items))
		for i, item := range list.Items {
			result[i] = item.Payload
		}
		return result, nil
	}
}

func getFields(idx resource.Index, s schema.Schema) graphql.Fields {
	flds := graphql.Fields{}
	// Iter fields
	for name, def := range s.Fields {
		if def.Hidden {
			continue
		}
		if _, ok := def.Validator.(*schema.Reference); ok {
			// Handled by addConnections to prevent dead loops.
		}
		var typ graphql.Output
		if def.Schema != nil {
			typ = graphql.NewObject(graphql.ObjectConfig{
				Name:   name,
				Fields: getFields(idx, *def.Schema),
			})
		} else {
			typ = getFType(def.Validator)
		}
		flds[name] = &graphql.Field{
			Description: def.Description,
			Type:        typ,
			Args:        getFArgs(def.Params),
			Resolve:     getFResolver(name, def),
		}
	}
	return flds
}

func getFArgs(p schema.Params) graphql.FieldConfigArgument {
	if p == nil {
		return nil
	}
	args := graphql.FieldConfigArgument{}
	for name, param := range p {
		args[name] = &graphql.ArgumentConfig{
			Description: param.Description,
			Type:        getFType(param.Validator),
		}
	}
	return args
}

// getFResolver returns a GraphQL field resolver for REST layer field handler.
func getFResolver(fieldName string, f schema.Field) graphql.FieldResolveFn {
	s, serialize := f.Validator.(schema.FieldSerializer)
	if !serialize && f.Handler == nil {
		return nil
	}
	return func(rp graphql.ResolveParams) (interface{}, error) {
		data, ok := rp.Source.(map[string]interface{})
		if !ok {
			return nil, nil
		}
		var err error
		val := data[fieldName]
		if f.Handler != nil {
			val, err = f.Handler(rp.Context, val, rp.Args)
		}
		if err == nil && serialize {
			val, err = s.Serialize(val)
		}
		return val, err
	}
}

// getFType translates a REST layer field type into GraphQL type.
func getFType(v schema.FieldValidator) graphql.Output {
	switch v.(type) {
	case *schema.String, schema.String:
		return graphql.String
	case *schema.Integer, schema.Integer:
		return graphql.Int
	case *schema.Float, schema.Float:
		return graphql.Float
	case *schema.Bool, schema.Bool:
		return graphql.Boolean
	default:
		return graphql.String
	}
}
