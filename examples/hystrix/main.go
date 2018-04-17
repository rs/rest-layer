// +build go1.7

package main

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/justinas/alice"
	"github.com/rs/rest-layer-hystrix"
	"github.com/rs/rest-layer/resource/testing/mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
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

	index.Bind("posts", post, restrix.Wrap("posts", mem.NewHandler()), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	// Create API HTTP handler for the resource graph
	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid API configuration")
	}

	// Setup logger
	c := alice.New()
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

	// Bind the API under the root path
	http.Handle("/", c.Then(api))

	// Configure hystrix commands
	hystrix.Configure(map[string]hystrix.CommandConfig{
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
	log.Info().Msg("Serving Hystrix metrics on http://localhost:8081")
	go http.ListenAndServe(net.JoinHostPort("", "8081"), hystrixStreamHandler)

	// Inject some fixtures
	fixtures := [][]string{
		{"POST", "/posts", `{"title": "First Post", "body": "This is my first post"}`},
		{"POST", "/posts", `{"title": "Second Post", "body": "This is my second post"}`},
		{"POST", "/posts", `{"title": "Third Post", "body": "This is my third post"}`},
	}
	for _, fixture := range fixtures {
		req, err := http.NewRequest(fixture[0], fixture[1], strings.NewReader(fixture[2]))
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
		w := httptest.NewRecorder()
		api.ServeHTTP(w, req)
		if w.Code >= 400 {
			log.Fatal().Msgf("Error returned for `%s %s`: %v", fixture[0], fixture[1], w)
		}
	}

	// Serve it
	log.Info().Msg("Serving API on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}
