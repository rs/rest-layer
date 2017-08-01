// +build go1.7

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/justinas/alice"
	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

// NOTE: this example show how to integrate REST Layer with JWT. No authentication is performed
// in this example. It is assumed that you are using a third party authentication system that
// generates JWT tokens with a user_id claim.

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

// NewJWTHandler parse and validates JWT token if present and store it in the net/context
func NewJWTHandler(users *resource.Resource, jwtKeyFunc jwt.Keyfunc) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := request.ParseFromRequest(r, request.OAuth2Extractor, jwtKeyFunc)
			if err == request.ErrNoTokenInRequest {
				// If no token is found, let REST Layer hooks decide if the resource is public or not
				next.ServeHTTP(w, r)
				return
			}
			if err != nil || !token.Valid {
				// Here you may want to return JSON error
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			claims := token.Claims.(jwt.MapClaims)
			userID, ok := claims["user_id"].(string)
			if !ok || userID == "" {
				// The provided token is malformed, user_id claim is missing
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			// Lookup the user by its id
			ctx := r.Context()
			user, err := users.Get(ctx, userID)
			if user != nil && err == resource.ErrUnauthorized {
				// Ignore unauthorized errors set by ourselves (see AuthResourceHook)
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
			// Store it into the request's context
			ctx = NewContextWithUser(ctx, user)
			// Add the user to log context (using zerolog)
			ctx = hlog.FromRequest(r).With().Interface("user_id", user.ID).Logger().WithContext(ctx)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// AuthResourceHook is a resource event handler that protect the resource from unauthorized users
type AuthResourceHook struct {
	UserField string
}

// OnFind implements resource.FindEventHandler interface
func (a AuthResourceHook) OnFind(ctx context.Context, q *query.Query) error {
	// Reject unauthorized users
	user, found := UserFromContext(ctx)
	if !found {
		return resource.ErrUnauthorized
	}
	// Add a predicate to the query to restrict to result on objects owned by this user
	q.Predicate = append(q.Predicate, query.Equal{Field: a.UserField, Value: user.ID})
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
func (a AuthResourceHook) OnClear(ctx context.Context, q *query.Query) error {
	// Reject unauthorized users
	user, found := UserFromContext(ctx)
	if !found {
		return resource.ErrUnauthorized
	}
	// Add a predicate to the query to restrict to impact of the clear on objects owned by this user
	q.Predicate = append(q.Predicate, query.Equal{Field: a.UserField, Value: user.ID})
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

var (
	jwtSecret = flag.String("jwt-secret", "secret", "The JWT secret passphrase")
)

func main() {
	flag.Parse()

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
			"id":       "jack",
			"name":     "Jack Sparrow",
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
		log.Fatal().Msgf("Invalid API configuration: %s", err)
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

	// Setup auth middleware
	jwtSecretBytes := []byte(*jwtSecret)
	c = c.Append(NewJWTHandler(users, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrInvalidKey
		}
		return jwtSecretBytes, nil
	}))

	// Bind the API under /
	http.Handle("/", c.Then(api))

	// Demo tokens
	jackToken := jwt.New(jwt.SigningMethodHS256)
	jackClaims := jackToken.Claims.(jwt.MapClaims)
	jackClaims["user_id"] = "jack"
	jackTokenString, err := jackToken.SignedString(jwtSecretBytes)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	johnToken := jwt.New(jwt.SigningMethodHS256)
	johnClaims := johnToken.Claims.(jwt.MapClaims)
	johnClaims["user_id"] = "john"
	johnTokenString, err := johnToken.SignedString(jwtSecretBytes)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Serve it
	fmt.Println("Serving API on http://localhost:8080")
	fmt.Printf("Your token secret is %q, change it with the `-jwt-secret' flag\n", *jwtSecret)
	fmt.Print("Play with tokens:\n" +
		"\n" +
		"- http :8080/posts access_token==" + johnTokenString + " title=\"John's post\"\n" +
		"- http :8080/posts access_token==" + johnTokenString + "\n" +
		"- http :8080/posts\n" +
		"\n" +
		"- http :8080/posts access_token==" + jackTokenString + " title=\"Jack's post\"\n" +
		"- http :8080/posts access_token==" + jackTokenString + "\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}
