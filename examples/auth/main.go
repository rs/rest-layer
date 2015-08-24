package main

import (
	"log"
	"net/http"
	"time"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"golang.org/x/net/context"
)

type myAuthMiddleware struct {
	userResource *resource.Resource
}

var (
	// Define a user resource schema
	user = schema.Schema{
		"id": schema.Field{
			Validator: &schema.String{
				MinLen: 2,
				MaxLen: 50,
			},
		},
		"name": schema.Field{
			Required:   true,
			Filterable: true,
			Validator: &schema.String{
				MaxLen: 150,
			},
		},
		"password": schema.Field{
			Required:  true,
			Validator: &schema.Password {
			MinLen: 6,
			Hide: true,
		},
		},
	}

	// Define a post resource schema
	post = schema.Schema{
		"id": schema.IDField,
		// Define a user field which references the user owning the post.
		// See bellow, the content of this field is enforced by the fact
		// that posts is a sub-resource of users.
		"user": schema.Field{
			Required:   true,
			Filterable: true,
			Validator: &schema.Reference{
				Path: "users",
			},
		},
		"title": schema.Field{
			Required: true,
			Validator: &schema.String{
				MaxLen: 150,
			},
		},
		"body": schema.Field{
			Validator: &schema.String{},
		},
	}
)

func (m myAuthMiddleware) Handle(ctx context.Context, r *http.Request, next rest.Next) (context.Context, int, http.Header, interface{}) {
	if u, p, ok := r.BasicAuth(); ok {
		// Lookup the user by its id
		lookup := resource.NewLookupWithQuery(schema.Query{
			schema.Equal{Field: "id", Value: u},
		})
		list, err := m.userResource.Find(ctx, lookup, 1, 1)
		if err != nil {
			// If user resource storage handler returned an error, stop the middleware chain
			return ctx, 0, nil, err
		}
		if len(list.Items) == 1 {
			user := list.Items[0]
			if schema.VerifyPassword(user.Payload["password"], []byte(p)) {
				// Get the current route from the context
				route, ok := rest.RouteFromContext(ctx)
				if ok {
					// If the current resource is "users", set the resource field to "id"
					// as user resource doesn't reference itself thru a "user" field.
					field := "user"
					if route.ResourcePath.Path() == "users" {
						field = "id"
					}
					// Prepent the resource path with the user resource
					route.PrependResourcePath(m.userResource, field, u)
					// Go the the next middleware
					return next(ctx)
				}
			}
		}
	}
	// Stop the middleware chain and return a 401 HTTP error
	headers := http.Header{}
	headers.Set("WWW-Authenticate", "Basic realm=\"API\"")
	return ctx, 401, headers, &rest.Error{401, "Please provide proper credentials", nil}
}

func main() {
	// Create a REST API resource index
	index := resource.NewIndex()

	// Bind user on /users
	users := index.Bind("users", resource.New(user, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	}))

	// Init the db with some users (user registration is not handled by this example)
	secret, _ := schema.Password{}.Validate("secret")
	users.Insert(context.Background(), []*resource.Item{
		&resource.Item{ID: "admin", Updated: time.Now(), ETag: "abcd", Payload: map[string]interface{}{
			"id":       "admin",
			"name":     "Dilbert",
			"password": secret,
		}},
		&resource.Item{ID: "john", Updated: time.Now(), ETag: "efgh", Payload: map[string]interface{}{
			"id":       "john",
			"name":     "John Doe",
			"password": secret,
		}},
	})

	// Bind post on /posts
	index.Bind("posts", resource.New(post, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	}))

	// Create API HTTP handler for the resource graph
	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatalf("Invalid API configuration: %s", err)
	}

	// Bind the authentication middleware
	api.Use(myAuthMiddleware{userResource: users})

	// Bind the API under /
	http.Handle("/", api)

	// Serve it
	log.Print("Serving API on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
