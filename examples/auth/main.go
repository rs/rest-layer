// +build go1.7

package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/xaccess"
	"github.com/rs/xlog"
)

// NOTE: this exemple demonstrates how to implement basic authentication/authorization with REST Layer.
// By no mean, we recommend to use basic authentication with your API. You can read more about auth
// best practices with REST Layer at http://rest-layer.io#authentication-and-authorization.

type key int

const userKey key = 0

// NewContextWithUser stores user into context
func NewContextWithUser(ctx context.Context, user *resource.Item) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext retrieves user from context
func UserFromContext(ctx context.Context) (*resource.Item, bool) {
	user, ok := ctx.Value(userKey).(*resource.Item)
	return user, ok
}

// NewBasicAuthHandler handles basic HTTP auth against the provided user resource
func NewBasicAuthHandler(users *resource.Resource) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if u, p, ok := r.BasicAuth(); ok {
				// Lookup the user by its id
				ctx := r.Context()
				user, err := users.Get(ctx, u)
				if user != nil && err == resource.ErrUnauthorized {
					// Ignore unauthorized errors set by ourselves
					err = nil
				}
				if err != nil {
					// If user resource storage handler returned an error, respond with an error
					if err == resource.ErrNotFound {
						http.Error(w, "Invalid credential", http.StatusForbidden)
					} else {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
				if schema.VerifyPassword(user.Payload["password"], []byte(p)) {
					// Store the auth user into the context for later use
					r = r.WithContext(NewContextWithUser(ctx, user))
					next.ServeHTTP(w, r)
					return
				}
			}
			// Stop the middleware chain and return a 401 HTTP error
			w.Header().Set("WWW-Authenticate", `Basic realm="API"`)
			http.Error(w, "Please provide proper credentials", http.StatusUnauthorized)
		})
	}
}

// AuthResourceHook is a resource event handler that protect the resource from unauthorized users
type AuthResourceHook struct {
	UserField string
}

// OnFind implements resource.FindEventHandler interface
func (a AuthResourceHook) OnFind(ctx context.Context, lookup *resource.Lookup, offset, limit int) error {
	// Reject unauthorized users
	user, found := UserFromContext(ctx)
	if !found {
		return resource.ErrUnauthorized
	}
	// Add a lookup condition to restrict to result on objects owned by this user
	lookup.AddQuery(schema.Query{
		schema.Equal{Field: a.UserField, Value: user.ID},
	})
	return nil
}

// OnGot implements resource.GotEventHandler interface
func (a AuthResourceHook) OnGot(ctx context.Context, item **resource.Item, err *error) {
	// Do not override existing errors
	if err != nil {
		return
	}
	// Reject unauthorized users
	user, found := UserFromContext(ctx)
	if !found {
		*err = resource.ErrUnauthorized
		return
	}
	// Check access right
	if u, found := (*item).Payload[a.UserField]; !found || u != user.ID {
		*err = resource.ErrNotFound
	}
	return
}

// OnInsert implements resource.InsertEventHandler interface
func (a AuthResourceHook) OnInsert(ctx context.Context, items []*resource.Item) error {
	// Reject unauthorized users
	user, found := UserFromContext(ctx)
	if !found {
		return resource.ErrUnauthorized
	}
	// Check access right
	for _, item := range items {
		if u, found := item.Payload[a.UserField]; found {
			if u != user.ID {
				return resource.ErrUnauthorized
			}
		} else {
			// If no user set for the item, set it to current user
			item.Payload[a.UserField] = user.ID
		}
	}
	return nil
}

// OnUpdate implements resource.UpdateEventHandler interface
func (a AuthResourceHook) OnUpdate(ctx context.Context, item *resource.Item, original *resource.Item) error {
	// Reject unauthorized users
	user, found := UserFromContext(ctx)
	if !found {
		return resource.ErrUnauthorized
	}
	// Check access right
	if u, found := original.Payload[a.UserField]; !found || u != user.ID {
		return resource.ErrUnauthorized
	}
	// Ensure user field is not altered
	if u, found := item.Payload[a.UserField]; !found || u != user.ID {
		return resource.ErrUnauthorized
	}
	return nil
}

// OnDelete implements resource.DeleteEventHandler interface
func (a AuthResourceHook) OnDelete(ctx context.Context, item *resource.Item) error {
	// Reject unauthorized users
	user, found := UserFromContext(ctx)
	if !found {
		return resource.ErrUnauthorized
	}
	// Check access right
	if item.Payload[a.UserField] != user.ID {
		return resource.ErrUnauthorized
	}
	return nil
}

// OnClear implements resource.ClearEventHandler interface
func (a AuthResourceHook) OnClear(ctx context.Context, lookup *resource.Lookup) error {
	// Reject unauthorized users
	user, found := UserFromContext(ctx)
	if !found {
		return resource.ErrUnauthorized
	}
	// Add a lookup condition to restrict to impact of the clear on objects owned by this user
	lookup.AddQuery(schema.Query{
		schema.Equal{Field: a.UserField, Value: user.ID},
	})
	return nil
}

var (
	// Define a user resource schema
	user = schema.Schema{
		Fields: schema.Fields{
			"id": {
				Validator: &schema.String{
					MinLen: 2,
					MaxLen: 50,
				},
			},
			"name": {
				Required:   true,
				Filterable: true,
				Validator: &schema.String{
					MaxLen: 150,
				},
			},
			"password": schema.PasswordField,
		},
	}

	// Define a post resource schema
	post = schema.Schema{
		Fields: schema.Fields{
			"id": schema.IDField,
			// Define a user field which references the user owning the post.
			// See bellow, the content of this field is enforced by the fact
			// that posts is a sub-resource of users.
			"user": {
				Required:   true,
				Filterable: true,
				Validator: &schema.Reference{
					Path: "users",
				},
				OnInit: func(ctx context.Context, value interface{}) interface{} {
					// If not set, set the user to currently logged user if any
					if value == nil {
						if user, found := UserFromContext(ctx); found {
							value = user.ID
						}
					}
					return value
				},
			},
			"title": {
				Required: true,
				Validator: &schema.String{
					MaxLen: 150,
				},
			},
			"body": {
				Validator: &schema.String{},
			},
		},
	}
)

func main() {
	// Create a REST API resource index
	index := resource.NewIndex()

	// Bind user on /users
	users := index.Bind("users", user, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	// Init the db with some users (user registration is not handled by this example)
	secret, _ := schema.Password{}.Validate("secret")
	users.Insert(context.Background(), []*resource.Item{
		{ID: "admin", Updated: time.Now(), ETag: "abcd", Payload: map[string]interface{}{
			"id":       "admin",
			"name":     "Dilbert",
			"password": secret,
		}},
		{ID: "john", Updated: time.Now(), ETag: "efgh", Payload: map[string]interface{}{
			"id":       "john",
			"name":     "John Doe",
			"password": secret,
		}},
	})

	// Bind post on /posts
	posts := index.Bind("posts", post, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	// Protect resources
	users.Use(AuthResourceHook{UserField: "id"})
	posts.Use(AuthResourceHook{UserField: "user"})

	// Create API HTTP handler for the resource graph
	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatalf("Invalid API configuration: %s", err)
	}

	// Setup logger
	c := alice.New()
	c = c.Append(xlog.NewHandler(xlog.Config{}))
	c = c.Append(xaccess.NewHandler())
	c = c.Append(xlog.RequestHandler("req"))
	c = c.Append(xlog.RemoteAddrHandler("ip"))
	c = c.Append(xlog.UserAgentHandler("ua"))
	c = c.Append(xlog.RefererHandler("ref"))
	c = c.Append(xlog.RequestIDHandler("req_id", "Request-Id"))
	resource.LoggerLevel = resource.LogLevelDebug
	resource.Logger = func(ctx context.Context, level resource.LogLevel, msg string, fields map[string]interface{}) {
		xlog.FromContext(ctx).OutputF(xlog.Level(level), 2, msg, fields)
	}

	// Setup auth middleware
	c = c.Append(NewBasicAuthHandler(users))

	// Bind the API under /
	http.Handle("/", c.Then(api))

	// Serve it
	log.Print("Serving API on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
