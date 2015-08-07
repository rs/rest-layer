/*
Package rest layer is a REST API framework heavily inspired by the excellent
Python Eve (http://python-eve.org).

It lets you automatically generate a comprehensive, customizable, and secure
REST API on top of any backend storage with no boiler plate code. You can focus
on your business logic now.

Implemented as a `net/http` middleware, it plays well with other middlewares like
CORS (http://github.com/rs/cors).

REST Layer is an opinionated framework. Unlike many web frameworks, you don't
directly control the routing. You just expose resources and sub-resources, the
framework automatically figures what routes to generate behind the scene.
You don't have to take care of the HTTP headers and response, JSON encoding, etc.
either. rest handles HTTP conditional requests, caching, integrity checking for
you. A powerful and extensible validation engine make sure that data comes
pre-validated to you resource handlers. Generic resource handlers for MongoDB and
other databases are also available so you have few to no code to write to make
the whole system work.
*/
package rest
