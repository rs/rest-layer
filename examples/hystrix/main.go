package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/xaccess"
	"github.com/rs/xhandler"
	"github.com/rs/xlog"
)

var (
	post = schema.Schema{
		Fields: schema.Fields{
			"id":      schema.IDField,
			"created": schema.CreatedField,
			"updated": schema.UpdatedField,
			"title":   {},
			"body":    {},
		},
	}
)

func main() {
	index := resource.NewIndex()

	index.Bind("posts", post, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
		Hystrix:      true,
	})

	// Create API HTTP handler for the resource graph
	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatalf("Invalid API configuration: %s", err)
	}

	// Setup logger
	c := xhandler.Chain{}
	c.UseC(xlog.NewHandler(xlog.Config{}))
	c.UseC(xaccess.NewHandler())

	// Bind the API under the root path
	http.Handle("/", c.Handler(api))

	// Configure hystrix commands
	hystrix.Configure(map[string]hystrix.CommandConfig{
		"posts.Get": {
			Timeout:               500,
			MaxConcurrentRequests: 200,
			ErrorPercentThreshold: 25,
		},
		"posts.MultiGet": {
			Timeout:               500,
			MaxConcurrentRequests: 200,
			ErrorPercentThreshold: 25,
		},
		"posts.Find": {
			Timeout:               1000,
			MaxConcurrentRequests: 100,
			ErrorPercentThreshold: 25,
		},
		"posts.Insert": {
			Timeout:               1000,
			MaxConcurrentRequests: 50,
			ErrorPercentThreshold: 25,
		},
		"posts.Update": {
			Timeout:               1000,
			MaxConcurrentRequests: 50,
			ErrorPercentThreshold: 25,
		},
		"posts.Delete": {
			Timeout:               1000,
			MaxConcurrentRequests: 10,
			ErrorPercentThreshold: 10,
		},
		"posts.Clear": {
			Timeout:               10000,
			MaxConcurrentRequests: 5,
			ErrorPercentThreshold: 10,
		},
	})

	// Start the metrics stream handler
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	log.Print("Serving Hystrix metrics on http://localhost:8081")
	go http.ListenAndServe(net.JoinHostPort("", "8081"), hystrixStreamHandler)

	// Inject some fixtures
	fixtures := [][]string{
		[]string{"POST", "/posts", `{"title": "First Post", "body": "This is my first post"}`},
		[]string{"POST", "/posts", `{"title": "Second Post", "body": "This is my second post"}`},
		[]string{"POST", "/posts", `{"title": "Third Post", "body": "This is my third post"}`},
	}
	for _, fixture := range fixtures {
		req, err := http.NewRequest(fixture[0], fixture[1], strings.NewReader(fixture[2]))
		if err != nil {
			log.Fatal(err)
		}
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		if w.Code >= 400 {
			log.Fatalf("Error returned for `%s %s`: %v", fixture[0], fixture[1], w)
		}
	}

	// Serve it
	log.Print("Serving API on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
