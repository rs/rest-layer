package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"golang.org/x/net/context"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/graphql"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/xaccess"
	"github.com/rs/xhandler"
	"github.com/rs/xlog"
)

var (
	user = schema.Schema{
		Description: "Defines user information",
		Fields: schema.Fields{
			"id": {
				Required:   true,
				ReadOnly:   true,
				Filterable: true,
				Sortable:   true,
				Validator: &schema.String{
					Regexp: "^[0-9a-z_-]{2,150}$",
				},
			},
			"created": schema.CreatedField,
			"updated": schema.UpdatedField,
			"name":    {},
			"admin": {
				Filterable: true,
				Validator:  &schema.Bool{},
			},
			"ip":       {Validator: &schema.IP{StoreBinary: true}},
			"password": schema.PasswordField,
		},
	}

	postFollower = schema.Schema{
		Description: "Link a post to its followers",
		Fields: schema.Fields{
			"id": schema.IDField,
			"post": {
				Validator: &schema.Reference{Path: "posts"},
			},
			"user": {
				Filterable: true,
				Sortable:   true,
				Validator:  &schema.Reference{Path: "users"},
			},
		},
	}

	post = schema.Schema{
		Description: "Defines a blog post",
		Fields: schema.Fields{
			"id":      schema.IDField,
			"created": schema.CreatedField,
			"updated": schema.UpdatedField,
			"user": {
				Validator: &schema.Reference{Path: "users"},
			},
			"thumbnail_url": {
				Description: "Resizable thumbnail URL for a post. Use width and height parameters to get a specific size.",
				Params: schema.Params{
					"width": {
						Description: "Change the width of the thumbnail to the value in pixels",
						Validator: schema.Integer{
							Boundaries: &schema.Boundaries{Max: 1000},
						},
					},
					"height": {
						Description: "Change the height of the thumbnail to the value in pixels",
						Validator: schema.Integer{
							Boundaries: &schema.Boundaries{Max: 1000},
						},
					},
				},
				// Appends a "w" and/or "h" query string parameter(s) to the value (URL) if width or height params passed
				Handler: func(ctx context.Context, value interface{}, params map[string]interface{}) (interface{}, error) {
					str, ok := value.(string)
					if !ok {
						return nil, errors.New("not a string")
					}
					sep := "?"
					if strings.IndexByte(str, '?') > 0 {
						sep = "&"
					}
					if width, found := params["width"]; found {
						str = fmt.Sprintf("%s%sw=%d", str, sep, width)
						sep = "&"
					}
					if height, found := params["height"]; found {
						str = fmt.Sprintf("%s%sy=%d", str, sep, height)
					}
					return str, nil
				},
			},
			"meta": {
				Schema: &schema.Schema{
					Fields: schema.Fields{
						"title": {},
						"body":  {},
					},
				},
			},
		},
	}
)

func main() {
	index := resource.NewIndex()

	users := index.Bind("users", user, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	users.Alias("admin", url.Values{"filter": []string{`{"admin": true}`}})

	posts := index.Bind("posts", post, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
	})

	posts.Bind("followers", "post", postFollower, mem.NewHandler(), resource.Conf{
		AllowedModes: resource.ReadWrite,
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
	c.UseC(xlog.RequestHandler("req"))
	c.UseC(xlog.RemoteAddrHandler("ip"))
	c.UseC(xlog.UserAgentHandler("ua"))
	c.UseC(xlog.RefererHandler("ref"))
	c.UseC(xlog.RequestIDHandler("req_id", "Request-Id"))
	resource.LoggerLevel = resource.LogLevelDebug
	resource.Logger = func(ctx context.Context, level resource.LogLevel, msg string, fields map[string]interface{}) {
		xlog.FromContext(ctx).OutputF(xlog.Level(level), 2, msg, fields)
	}

	// Bind the API under /api/ path
	http.Handle("/api/", http.StripPrefix("/api/", c.Handler(api)))

	// Create and bind the graphql endpoint
	graphql, err := graphql.NewHandler(index)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/graphql", c.Handler(graphql))
	http.HandleFunc("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
  <style>
    html, body {height: 100%; margin: 0; overflow: hidden; width: 100%;}
  </style>
  <link href="//cdn.jsdelivr.net/graphiql/0.4.9/graphiql.css" rel="stylesheet" />
  <script src="//cdn.jsdelivr.net/fetch/0.9.0/fetch.min.js"></script>
  <script src="//cdn.jsdelivr.net/react/0.14.7/react.min.js"></script>
  <script src="//cdn.jsdelivr.net/react/0.14.7/react-dom.min.js"></script>
  <script src="//cdn.jsdelivr.net/graphiql/0.4.9/graphiql.min.js"></script>
</head>
<body>
  <script>
    // Collect the URL parameters
    var parameters = {};
    window.location.search.substr(1).split('&').forEach(function (entry) {
      var eq = entry.indexOf('=');
      if (eq >= 0) {
        parameters[decodeURIComponent(entry.slice(0, eq))] =
          decodeURIComponent(entry.slice(eq + 1));
      }
    });

    // Produce a Location query string from a parameter object.
    function locationQuery(params) {
      return '/graphql?' + Object.keys(params).map(function (key) {
        return encodeURIComponent(key) + '=' +
          encodeURIComponent(params[key]);
      }).join('&');
    }

    // Derive a fetch URL from the current URL, sans the GraphQL parameters.
    var graphqlParamNames = {
      query: true,
      variables: true,
      operationName: true
    };

    var otherParams = {};
    for (var k in parameters) {
      if (parameters.hasOwnProperty(k) && graphqlParamNames[k] !== true) {
        otherParams[k] = parameters[k];
      }
    }
    var fetchURL = locationQuery(otherParams);

    // Defines a GraphQL fetcher using the fetch API.
    function graphQLFetcher(graphQLParams) {
      return fetch(fetchURL, {
        method: 'post',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(graphQLParams),
        credentials: 'include',
      }).then(function (response) {
        return response.text();
      }).then(function (responseBody) {
        try {
          return JSON.parse(responseBody);
        } catch (error) {
          return responseBody;
        }
      });
    }

    // When the query and variables string is edited, update the URL bar so
    // that it can be easily shared.
    function onEditQuery(newQuery) {
      parameters.query = newQuery;
      updateURL();
    }

    function onEditVariables(newVariables) {
      parameters.variables = newVariables;
      updateURL();
    }

    function updateURL() {
      history.replaceState(null, null, locationQuery(parameters));
    }

    // Render <GraphiQL /> into the body.
    React.render(
      React.createElement(GraphiQL, {
        fetcher: graphQLFetcher,
        onEditQuery: onEditQuery,
        onEditVariables: onEditVariables,
		defaultQuery: "{\
  postsList{\
    i: id,\
    m: meta{\
      t: title,\
      b: body},\
    thumb_small_url: thumbnail_url(height:80)\
  }\
}",
      }),
      document.body
    );
  </script>
</body>
</html>`))
	})

	// Inject some fixtures
	fixtures := [][]string{
		{"PUT", "/users/johndoe", `{"name": "John Doe", "ip": "1.2.3.4", "password": "secret", "admin": true}`},
		{"PUT", "/users/fan1", `{"name": "Fan 1", "ip": "1.2.3.4", "password": "secret"}}`},
		{"PUT", "/users/fan2", `{"name": "Fan 2", "ip": "1.2.3.4", "password": "secret"}}`},
		{"PUT", "/users/fan3", `{"name": "Fan 3", "ip": "1.2.3.4", "password": "secret"}}`},
		{"PUT", "/users/fan4", `{"name": "Fan 4", "ip": "1.2.3.4", "password": "secret"}}`},
		{"PUT", "/posts/ar5qrgukj5l7a6eq2ps0",
			`{
				"user": "johndoe",
				"thumbnail_url": "http://dom.com/image.png",
				"meta": {
					"title": "First Post",
					"body": "This is my first post"
				}
			}`},
		{"POST", "/posts/ar5qrgukj5l7a6eq2ps0/followers", `{"user": "fan1"}`},
		{"POST", "/posts/ar5qrgukj5l7a6eq2ps0/followers", `{"user": "fan2"}`},
		{"POST", "/posts/ar5qrgukj5l7a6eq2ps0/followers", `{"user": "fan3"}`},
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
	log.Print("Visit http://localhost:8080/graphiql for a GraphiQL UI")
	log.Println("Play with (httpie):\n",
		"- http :8080/graphql query=='{postsList{id,thumb_s_url:thumbnail_url(height:80)}}'\n",
		"- http :8080/graphql query=='{postsList{i:id,m:meta{t:title, b:body},thumb_small_url:thumbnail_url(height:80)}}'\n",
		"- http :8080/graphql query=='{postsList{id,meta{title},user{id,name}}}'\n",
		"- http :8080/graphql query=='{posts(id:\"ar5qrgukj5l7a6eq2ps0\"){followers{post{id,meta{title}},user{id,name}}}}'\n",
		"- http :8080/graphql query=='{posts(id:\"ar5qrgukj5l7a6eq2ps0\"){id,meta{title},followers(limit:2){user{id,name}}}}'")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
