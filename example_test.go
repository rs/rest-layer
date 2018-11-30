package restlayer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/rs/cors"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/resource/testing/mem"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
)

// ResponseRecorder extends http.ResponseWriter with the ability to capture
// the status and number of bytes written
type ResponseRecorder struct {
	http.ResponseWriter

	statusCode int
	length     int
}

// NewResponseRecorder returns a ResponseRecorder that wraps w.
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

// Write writes b to the underlying response writer and stores how many bytes
// have been written.
func (w *ResponseRecorder) Write(b []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(b)
	w.length += n
	return
}

// WriteHeader stores and writes the HTTP status code.
func (w *ResponseRecorder) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// StatusCode returns the status-code written to the response or 200 (OK).
func (w *ResponseRecorder) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := NewResponseRecorder(w)

		next.ServeHTTP(rec, r)
		status := rec.StatusCode()
		length := rec.length

		// In this example we use the standard library logger. Structured logs
		// may prove more parsable in a production environment.
		log.Printf("D! Served HTTP Request %s %s with Response %d %s [%d bytes] in %d ms",
			r.Method,
			r.URL,
			status,
			http.StatusText(status),
			length,
			time.Since(start).Nanoseconds()/1e6,
		)
	})
}

var logLevelPrefixes = map[resource.LogLevel]string{
	resource.LogLevelFatal: "E!",
	resource.LogLevelError: "E!",
	resource.LogLevelWarn:  "W!",
	resource.LogLevelInfo:  "I!",
	resource.LogLevelDebug: "D!",
}

func Example() {
	// Configure a log-addapter for the resource pacakge.
	resource.LoggerLevel = resource.LogLevelDebug
	resource.Logger = func(ctx context.Context, level resource.LogLevel, msg string, fields map[string]interface{}) {
		fmt.Printf("%s %s %v", logLevelPrefixes[level], msg, fields)
	}

	var (
		// Define a user resource schema
		user = schema.Schema{
			Fields: schema.Fields{
				"id": {
					Required: true,
					// When a field is read-only, on default values or hooks can
					// set their value. The client can't change it.
					ReadOnly: true,
					// This is a field hook called when a new user is created.
					// The schema.NewID hook is a provided hook to generate a
					// unique id when no value is provided.
					OnInit: schema.NewID,
					// The Filterable and Sortable allows usage of filter and sort
					// on this field in requests.
					Filterable: true,
					Sortable:   true,
					Validator: &schema.String{
						Regexp: "^[0-9a-f]{32}$",
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
					Validator: &schema.Reference{
						Path: "users",
					},
				},
				"public": {
					Filterable: true,
					Validator:  &schema.Bool{},
				},
				// Sub-documents are handled via a sub-schema
				"meta": {
					Schema: &schema.Schema{
						Fields: schema.Fields{
							"title": {
								Required: true,
								Validator: &schema.String{
									MaxLen: 150,
								},
							},
							"body": {
								Validator: &schema.String{
									MaxLen: 100000,
								},
							},
						},
					},
				},
			},
		}
	)

	// Create a REST API root resource.
	index := resource.NewIndex()

	// Add a resource on /users[/:user_id]
	users := index.Bind("users", user, mem.NewHandler(), resource.Conf{
		// We allow all REST methods.
		// (rest.ReadWrite is a shortcut for []rest.Mode{Create, Read, Update, Delete, List})
		AllowedModes: resource.ReadWrite,
	})

	// Bind a sub resource on /users/:user_id/posts[/:post_id]
	// and reference the user on each post using the "user" field of the posts resource.
	posts := users.Bind("posts", "user", post, mem.NewHandler(), resource.Conf{
		// Posts can only be read, created and deleted, not updated
		AllowedModes: []resource.Mode{resource.Read, resource.List, resource.Create, resource.Delete},
	})

	// Add a friendly alias to public posts.
	// (equivalent to /users/:user_id/posts?filter={"public":true})
	posts.Alias("public", url.Values{"filter": []string{"{\"public\"=true}"}})

	// Create API HTTP handler for the resource graph.
	var api http.Handler
	api, err := rest.NewHandler(index)
	if err != nil {
		log.Printf("E! Invalid API configuration: %s", err)
		os.Exit(1)
	}

	// Add CORS support with passthrough option on so rest-layer can still
	// handle OPTIONS method.
	api = cors.New(cors.Options{OptionsPassthrough: true}).Handler(api)

	// Wrap the api & cors handler with an access log middleware.
	api = AccessLog(api)

	// Bind the API under the /api/ path.
	http.Handle("/api/", http.StripPrefix("/api/", api))

	// Serve it.
	log.Printf("I! Serving API on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("E! Cannot serve API: %s", err)
		os.Exit(1)
	}
}
