package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/rs/cors"
	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
)

var (
	// Define a user resource schema
	user = schema.Schema{
		"id": schema.Field{
			Required: true,
			// When a field is read-only, on default values or hooks can
			// set their value. The client can't change it.
			ReadOnly: true,
			// This is a field hook called when a new user is created.
			// The schema.NewID hook is a provided hook to generate a
			// unique id when no value is provided.
			OnInit: &schema.NewID,
			// The Filterable and Sortable allows usage of filter and sort
			// on this field in requests.
			Filterable: true,
			Sortable:   true,
			Validator: &schema.String{
				Regexp: "^[0-9a-f]{24}$",
			},
		},
		"created": schema.Field{
			Required:   true,
			ReadOnly:   true,
			Filterable: true,
			Sortable:   true,
			OnInit:     &schema.Now,
			Validator:  &schema.Time{},
		},
		"updated": schema.Field{
			Required:   true,
			ReadOnly:   true,
			Filterable: true,
			Sortable:   true,
			OnInit:     &schema.Now,
			// The OnUpdate hook is called when the item is edited. Here we use
			// provided Now hook which just return the current time.
			OnUpdate:  &schema.Now,
			Validator: &schema.Time{},
		},
		// Define a name field as required with a string validator
		"name": schema.Field{
			Required:   true,
			Filterable: true,
			Validator: &schema.String{
				MaxLen: 150,
			},
		},
	}

	// Define a post resource schema
	post = schema.Schema{
		// schema.*Field are shortcuts for common fields (identical to users' same fields)
		"id":      schema.IDField,
		"created": schema.CreatedField,
		"updated": schema.UpdatedField,
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
		"published": schema.Field{
			Filterable: true,
			Validator:  &schema.Bool{},
		},
		// Sub-documents are handled via a sub-schema
		"meta": schema.Field{
			Schema: &schema.Schema{
				"title": schema.Field{
					Required: true,
					Validator: &schema.String{
						MaxLen: 150,
					},
				},
				"body": schema.Field{
					// Dependency defines that body field can't be changed if
					// the published field is not "false".
					Dependency: schema.Q("{\"published\": false}"),
					Validator: &schema.String{
						MaxLen: 100000,
					},
				},
			},
		},
	}
)

func main() {
	// Create a REST API resource index
	index := resource.NewIndex()

	// Add a resource on /users[/:user_id]
	users := index.Bind("users", resource.New(user, mem.NewHandler(), resource.Conf{
		// We allow all REST methods
		// (rest.ReadWrite is a shortcut for []resource.Mode{resource.Create, resource.Read, resource.Update, resource.Delete, resource,List})
		AllowedModes: resource.ReadWrite,
	}))

	// Bind a sub resource on /users/:user_id/posts[/:post_id]
	// and reference the user on each post using the "user" field of the posts resource.
	posts := users.Bind("posts", "user", resource.New(post, mem.NewHandler(), resource.Conf{
		// Posts can only be read, created and deleted, not updated
		AllowedModes: []resource.Mode{resource.Read, resource.List, resource.Create, resource.Delete},
	}))

	// Add a friendly alias to public posts
	// (equivalent to /users/:user_id/posts?filter={"published":true})
	posts.Alias("public", url.Values{"filter": []string{"{\"published\":true}"}})

	// Create API HTTP handler for the resource graph
	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatalf("Invalid API configuration: %s", err)
	}

	// Add cors support
	h := cors.New(cors.Options{OptionsPassthrough: true}).Handler(api)

	// Bind the API under /api/ path
	http.Handle("/api/", http.StripPrefix("/api/", h))

	// Serve it
	log.Print("Serving API on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
