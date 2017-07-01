/*
Package restlayer is an API framework heavily inspired by the excellent Python
Eve (http://python-eve.org/). It helps you create a comprehensive, customizable,
and secure REST (graph) API on top of pluggable backend storages with no boiler
plate code so can focus on your business logic.

Implemented as a net/http middleware, it plays well with other middleware like
CORS (http://github.com/rs/cors) and is net/context aware thanks to xhandler.

REST Layer is an opinionated framework. Unlike many API frameworks, you don’t
directly control the routing and you don’t have to write handlers. You just
define resources and sub-resources with a schema, the framework automatically
figures out what routes to generate behind the scene. You don’t have to take
care of the HTTP headers and response, JSON encoding, etc. either. REST layer
handles HTTP conditional requests, caching, integrity checking for you.

A powerful and extensible validation engine make sure that data comes
pre-validated to your custom storage handlers. Generic resource handlers for
MongoDB (http://github.com/rs/rest-layer-mongo), ElasticSearch
(http://github.com/rs/rest-layer-es) and other databases are also available so
you have few to no code to write to make the whole system work.

Moreover, REST Layer let you create a graph API by linking resources between
them. Thanks to its advanced field selection syntax (and coming support of
GraphQL), you can gather resources and their dependencies in a single request,
saving you from costly network roundtrips.

REST Layer is composed of several sub-packages:

 - rest: Holds the `net/http` handler responsible for the implementation of the
   RESTful API.
 - graphql: Holds a `net/http` handler to expose the API using the GraphQL
   protocol.
 - schema: Provides a validation framework for the API resources.
 - resource: Defines resources, manages the resource graph and manages the
   interface with resource storage handler.

See https://github.com/rs/rest-layer/blob/master/README.md for full REST Layer
documentation.
*/
package restlayer
