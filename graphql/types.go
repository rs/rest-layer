package graphql

import (
	"log"

	"github.com/graphql-go/graphql"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
)

type types map[string]*graphql.Object

// getObjectType returns a graphql object type definition from a REST layer schema
func (t types) getObjectType(idx resource.Index, name string, s schema.Schema) *graphql.Object {
	// Memoize types by their name so we don't create several instance of the same resource
	o := t[name]
	if o == nil {
		o = graphql.NewObject(graphql.ObjectConfig{
			Name:        name,
			Description: s.Description,
			Fields:      t.getFields(idx, s),
		})
		t[name] = o
	}
	return o
}

func (t types) getFields(idx resource.Index, s schema.Schema) graphql.Fields {
	flds := graphql.Fields{}
	for name, def := range s.Fields {
		if def.Schema != nil {
			flds[name] = &graphql.Field{
				Type: graphql.NewObject(graphql.ObjectConfig{
					Name:        name,
					Description: def.Description,
					Fields:      t.getFields(idx, *def.Schema),
				}),
			}
		} else if ref, ok := def.Validator.(*schema.Reference); ok {
			r, found := idx.GetResource(ref.Path, nil)
			if !found {
				log.Panicf("resource reference not found: %s", ref.Path)
			}
			flds[name] = t.getSubQuery(idx, r, name)
			flds[name].Description = def.Description
		} else {
			flds[name] = &graphql.Field{
				Description: def.Description,
				Type:        getFType(def.Validator),
				Args:        getFArgs(def.Params),
				Resolve:     getFResolver(name, def.Handler),
			}
		}
		// TODO: add sub-resources as fields
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

func getFResolver(fieldName string, h schema.FieldHandler) graphql.FieldResolveFn {
	if h == nil {
		return nil
	}
	return func(rp graphql.ResolveParams) (interface{}, error) {
		data, ok := rp.Source.(map[string]interface{})
		if !ok {
			return nil, nil
		}
		return h(data[fieldName], rp.Args)
	}
}

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
