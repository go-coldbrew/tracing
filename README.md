<!-- Code generated by gomarkdoc. DO NOT EDIT -->

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/go-coldbrew/tracing)

# tracing

```go
import "github.com/go-coldbrew/tracing"
```

Package tracing is a library that provides distributed tracing to Go applications. It offers features such as collecting performance data of an application, identifying where requests are spending most of their time, and segmenting requests. It supports exporting traces to 3rd\-party services such as Jaeger, Zipkin, Opentelemetry, and NewRelic. Go\-Coldbrew Tracing helps developers quickly identify issues and take corrective action when performance bottlenecks occur.

## Index

- [func ClientSpan(operationName string, ctx context.Context) (context.Context, opentracing.Span)](<#func-clientspan>)
- [func CloneContextValues(parent context.Context) context.Context](<#func-clonecontextvalues>)
- [func GRPCTracingSpan(operationName string, ctx context.Context) context.Context](<#func-grpctracingspan>)
- [func MergeContextValues(parent context.Context, main context.Context) context.Context](<#func-mergecontextvalues>)
- [func MergeParentContext(parent context.Context, main context.Context) context.Context](<#func-mergeparentcontext>)
- [func NewContextWithParentValues(parent context.Context) context.Context](<#func-newcontextwithparentvalues>)
- [type Span](<#type-span>)
  - [func NewDatastoreSpan(ctx context.Context, datastore, operation, collection string) (Span, context.Context)](<#func-newdatastorespan>)
  - [func NewExternalSpan(ctx context.Context, name string, url string) (Span, context.Context)](<#func-newexternalspan>)
  - [func NewHTTPExternalSpan(ctx context.Context, name string, url string, hdr http.Header) (Span, context.Context)](<#func-newhttpexternalspan>)
  - [func NewInternalSpan(ctx context.Context, name string) (Span, context.Context)](<#func-newinternalspan>)


## func [ClientSpan](<https://github.com/go-coldbrew/tracing/blob/main/tracing.go#L227>)

```go
func ClientSpan(operationName string, ctx context.Context) (context.Context, opentracing.Span)
```

ClientSpan starts a new client span linked to the existing spans if any are found in the context. The returned context should be used in place of the original

## func [CloneContextValues](<https://github.com/go-coldbrew/tracing/blob/main/context.go#L24>)

```go
func CloneContextValues(parent context.Context) context.Context
```

CloneContextValues clones a given context values and returns a new context obj which is not affected by Cancel, Deadline etc Deprecated: The function name is a bit confusing, use CloneContextValues instead

## func [GRPCTracingSpan](<https://github.com/go-coldbrew/tracing/blob/main/tracing.go#L245>)

```go
func GRPCTracingSpan(operationName string, ctx context.Context) context.Context
```

GRPCTracingSpan starts a new client span linked to the existing spans if any are found in the context. The returned context should be used in place of the original

## func [MergeContextValues](<https://github.com/go-coldbrew/tracing/blob/main/context.go#L45>)

```go
func MergeContextValues(parent context.Context, main context.Context) context.Context
```

MergeContextValues merged the given main context with a parent context, Cancel/Deadline etc are used from the main context and values are looked in both the contexts can be use to merge a parent context with a new context, the new context will have the values from both the contexts

## func [MergeParentContext](<https://github.com/go-coldbrew/tracing/blob/main/context.go#L39>)

```go
func MergeParentContext(parent context.Context, main context.Context) context.Context
```

MergeParentContext merged the given main context with a parent context, Cancel/Deadline etc are used from the main context and values are looked in both the contexts Deprecated: The function name is a bit confusing, use MergeContextValues instead

## func [NewContextWithParentValues](<https://github.com/go-coldbrew/tracing/blob/main/context.go#L30>)

```go
func NewContextWithParentValues(parent context.Context) context.Context
```

NewContextWithParentValues clones a given context values and returns a new context obj which is not affected by Cancel, Deadline etc can be used to pass context values to a new context which is not affected by the parent context cancel/deadline etc from parent

## type [Span](<https://github.com/go-coldbrew/tracing/blob/main/tracing.go#L19-L30>)

Span defines an interface for implementing a tracing span This is used to abstract the underlying tracing implementation, currently using opentracing/opentelemetry and newrelic tracing libraries for implementation

```go
type Span interface {
    // End ends the span, can also use Finish()
    End()
    // Finish ends the span, can also use End()
    Finish()
    // SetTag sets a tag on the span, can be used to add custom attributes
    SetTag(key string, value interface{})
    // SetQuery sets the query on the span, can be used to add query for datastore spans
    SetQuery(query string)
    // SetError sets the error on the span
    SetError(err error) error
}
```

### func [NewDatastoreSpan](<https://github.com/go-coldbrew/tracing/blob/main/tracing.go#L126>)

```go
func NewDatastoreSpan(ctx context.Context, datastore, operation, collection string) (Span, context.Context)
```

NewDatastoreSpan starts a span for tracing data store actions This is used to trace actions against a data store, for example, a database query or a redis call

### func [NewExternalSpan](<https://github.com/go-coldbrew/tracing/blob/main/tracing.go#L178>)

```go
func NewExternalSpan(ctx context.Context, name string, url string) (Span, context.Context)
```

NewExternalSpan starts a span for tracing external actions This is used to trace actions against an external service, for example, a call to another service or a call to an external API

### func [NewHTTPExternalSpan](<https://github.com/go-coldbrew/tracing/blob/main/tracing.go#L185>)

```go
func NewHTTPExternalSpan(ctx context.Context, name string, url string, hdr http.Header) (Span, context.Context)
```

NewHTTPExternalSpan starts a span for tracing external HTTP actions This is used to trace actions against an external service, for example, a call to another service or a call to an external API It also adds the HTTP headers to the span so that the external service can trace the call back to this service if needed

### func [NewInternalSpan](<https://github.com/go-coldbrew/tracing/blob/main/tracing.go#L109>)

```go
func NewInternalSpan(ctx context.Context, name string) (Span, context.Context)
```

NewInternalSpan starts a span for tracing internal actions This is used to trace actions within the same service, for example, a function call within the same service



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
