# REST Layer Memory backend [![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/rs/rest-layer/resource/testing/mem) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/rest-layer-mem/master/LICENSE) [![build](https://img.shields.io/travis/rs/rest-layer-mem.svg?style=flat)](https://travis-ci.org/rs/rest-layer-mem)

This REST Layer resource storage backend stores data in memory with no persistence. This package is provided as an implementation example and a test backend to be used for testing only.

**DO NOT USE THIS IN PRODUCTION.**

## Usage

Simply create a memory resource handler per resource:

```go
import "github.com/rs/rest-layer/resource/testing/mem"
```

```go
index.Bind("foo", foo, mem.NewHandler(), resource.DefaultConf)
```

## Latency Simulation

As local memory access is very fast, this handler is not very useful when it comes to working with latency related issues. This handler allows you to simulate latency by setting an artificial delay:

```go
root.Bind("foo", resource.NewResource(foo, mem.NewSlowHandler(5*time.Second), resource.DefaultConf)
```

With this configuration, the memory handler will pause 5 seconds before processing every request. If the passed `net/context` is canceled during that wait, the handler won't process the request and return the appropriate `rest.Error` as specified in the REST Layer [storage handler implementation doc](https://github.com/rs/rest-layer#data-storage-handler).
