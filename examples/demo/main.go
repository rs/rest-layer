// +build go1.7

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/cors"
	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

var (
	// Define a user resource schema
	user = schema.Schema{
		Fields: schema.Fields{
			"id": {
				Required: true,
				// The Filterable and Sortable allows usage of filter and sort
				// on this field in requests.
				Filterable: true,
				Sortable:   true,
				Validator: &schema.String{
					Regexp: "^[0-9a-z]{2,20}$",
				},
			},
			"created": {
				Required:   true,
				ReadOnly:   true,
				Filterable: true,
				Sortable:   true,
				OnInit:     schema.Now,
				Validator:  &schema.Time{},
			},
			"updated": {
				Required:   true,
				ReadOnly:   true,
				Filterable: true,
				Sortable:   true,
				OnInit:     schema.Now,
				// The OnUpdate hook is called when the item is edited. Here we use
				// provided Now hook which just return the current time.
				OnUpdate:  schema.Now,
				Validator: &schema.Time{},
			},
			// Define a name field as required with a string validator
			"name": {
				Required:   true,
				Filterable: true,
				Validator: &schema.String{
					MaxLen: 150,
				},
			},
		},
	}

	// Define a post resource schema
	post = schema.Schema{
		Fields: schema.Fields{
			// schema.*Field are shortcuts for common fields (identical to users' same fields)
			"id":      schema.IDField,
			"created": schema.CreatedField,
			"updated": schema.UpdatedField,
			// Define a user field which references the user owning the post.
			// See bellow, the content of this field is enforced by the fact
			// that posts is a sub-resource of users.
			"user": {
				Required:   true,
				Filterable: true,
				ReadOnly:   true,
				Validator: &schema.Reference{
					Path: "users",
				},
			},
			"published": {
				Filterable: true,
				Default:    false,
				Validator:  &schema.Bool{},
			},
			"title": {
				Required: true,
				Validator: &schema.String{
					MaxLen: 150,
				},
				// Dependency defines that body field can't be changed if
				// the published field is not "false".
				Dependency: query.MustParsePredicate(`{published: false}`),
			},
			"body": {
				Validator: &schema.String{
					MaxLen: 100000,
				},
				Dependency: query.MustParsePredicate(`{published: false}`),
			},
		},
	}
)

func main() {
	// Create a REST API resource index
	index := resource.NewIndex()

	// Add a resource on /users[/:user_id]
	users := index.Bind("users", user, mem.NewHandler(), resource.Conf{
		// We allow all REST methods
		// (rest.ReadWrite is a shortcut for []resource.Mode{resource.Create, resource.Read, resource.Update, resource.Delete, resource,List})
		AllowedModes: resource.ReadWrite,
	})

	// Bind a sub resource on /users/:user_id/posts[/:post_id]
	// and reference the user on each post using the "user" field of the posts resource.
	posts := users.Bind("posts", "user", post, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	// Add a friendly alias to public posts
	// (equivalent to /users/:user_id/posts?filter={"published":true})
	posts.Alias("public", url.Values{"filter": []string{"{\"published\":true}"}})

	// Create API HTTP handler for the resource graph
	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatal().Msgf("Invalid API configuration: %s", err)
	}

	c := alice.New()

	// Install a logger
	c = c.Append(hlog.NewHandler(log.With().Logger()))
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	c = c.Append(hlog.RequestHandler("req"))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("ua"))
	c = c.Append(hlog.RefererHandler("ref"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))
	resource.LoggerLevel = resource.LogLevelDebug
	resource.Logger = func(ctx context.Context, level resource.LogLevel, msg string, fields map[string]interface{}) {
		zerolog.Ctx(ctx).WithLevel(zerolog.Level(level)).Fields(fields).Msg(msg)
	}

	// Add CORS support with passthrough option on so rest-layer can still
	// handle OPTIONS method
	c = c.Append(cors.New(cors.Options{OptionsPassthrough: true}).Handler)

	// Bind the API under /api/ path
	http.Handle("/api/", http.StripPrefix("/api/", c.Then(api)))

	// Serve it
	fmt.Println("Serving API on http://localhost:8080")
	fmt.Println(`
Create a user:

	http PUT :8080/api/users/john name="John Doe"

Create a post for that user:

	http :8080/api/users/john/posts title="First Post" body="Lorem ipsum"

Edit the post:

	http PATCH :8080/api/users/john/posts/<post_id> body="Final body"

Publish:

	http PATCH :8080/api/users/john/posts/<post_id> published:=true

Once published, title and body can't be changed:

	http PATCH :8080/api/users/john/posts/<post_id> body="Final body"
	# returns 422

Get the post plus user name:

	http :8080/api/users/john/posts/<post_id> fields=='title,body,user{name}'
`)
	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}
