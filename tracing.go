package tracing

import (
	"context"
	"encoding/base64"
	"net/http"
	"runtime/trace"
	"strings"

	nrutil "github.com/go-coldbrew/tracing/newrelic"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
	opentracing "github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/metadata"
)

// Span defines an interface for implementing a tracing span
// This is used to abstract the underlying tracing implementation, currently using opentracing/opentelemetry and newrelic tracing libraries for implementation
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

type tracingSpan struct {
	openSpan        opentracing.Span
	datastore       bool
	external        bool
	dataSegment     newrelic.DatastoreSegment
	externalSegment newrelic.ExternalSegment
	segment         newrelic.Segment
	runtimeRegion   *trace.Region
	txn             *newrelic.Transaction
}

func (span *tracingSpan) End() {
	if span == nil {
		// dont panic when called against a nil span
		return
	}
	span.openSpan.Finish()

	if span.datastore {
		span.dataSegment.End()
	} else if span.external {
		span.externalSegment.End()
	} else {
		span.segment.End()
	}

	if span.txn != nil {
		span.txn.End()
	}

	span.runtimeRegion.End()
}

func (span *tracingSpan) Finish() {
	span.End()
}

func (span *tracingSpan) SetTag(key string, value interface{}) {
	if span == nil {
		// dont panic when called against a nil span
		return
	}
	span.openSpan.SetTag(key, value)
	if span.datastore {
		span.dataSegment.AddAttribute(key, value)
	} else if span.external {
		span.externalSegment.AddAttribute(key, value)
	} else {
		span.segment.AddAttribute(key, value)
	}
}

func (span *tracingSpan) SetQuery(query string) {
	if span == nil {
		// dont panic when called against a nil span
		return
	}
	span.openSpan.SetTag("query", query)
	if span.datastore {
		span.dataSegment.ParameterizedQuery = query
	}
}

func (span *tracingSpan) SetError(err error) error {
	if span == nil || err == nil {
		// dont panic when called against a nil span
		return err
	}
	span.openSpan.SetTag("error", "true")
	span.openSpan.SetTag("errorDetails", err.Error())
	if span.datastore {
		span.dataSegment.AddAttribute("error", err)
	} else if span.external {
		span.externalSegment.AddAttribute("error", err)
	} else {
		span.segment.AddAttribute("error", err)
	}
	return err
}

// NewInternalSpan starts a span for tracing internal actions
// This is used to trace actions within the same service, for example, a function call within the same service
func NewInternalSpan(ctx context.Context, name string) (Span, context.Context) {
	zip, ctx := opentracing.StartSpanFromContext(ctx, name)

	txnStarted := false
	txn := nrutil.GetNewRelicTransactionFromContext(ctx)
	if txn == nil {
		txnStarted = true
		ctx = nrutil.StartNRTransaction(name, ctx, nil, nil)
		txn = nrutil.GetNewRelicTransactionFromContext(ctx)
	}

	seg := newrelic.Segment{
		StartTime: txn.StartSegmentNow(),
		Name:      name,
	}
	reg := trace.StartRegion(ctx, name)
	span := &tracingSpan{
		openSpan:      zip,
		segment:       seg,
		runtimeRegion: reg,
	}
	if txnStarted {
		span.txn = txn
	}
	return span, ctx
}

// NewDatastoreSpan starts a span for tracing data store actions
// This is used to trace actions against a data store, for example, a database query or a redis call
func NewDatastoreSpan(ctx context.Context, datastore, operation, collection string) (Span, context.Context) {
	name := operation
	if !strings.HasPrefix(name, datastore) {
		name = datastore + name
	}
	zip, ctx := opentracing.StartSpanFromContext(ctx, name)
	zip.SetTag("store", datastore)
	zip.SetTag("collection", collection)
	zip.SetTag("operation", operation)

	txnStarted := false
	txn := nrutil.GetNewRelicTransactionFromContext(ctx)
	if txn == nil {
		txnStarted = true
		ctx = nrutil.StartNRTransaction(datastore+":"+operation+":"+collection, ctx, nil, nil)
		txn = nrutil.GetNewRelicTransactionFromContext(ctx)
	}

	seg := newrelic.DatastoreSegment{
		StartTime:  txn.StartSegmentNow(),
		Product:    newrelic.DatastoreProduct(datastore),
		Operation:  operation,
		Collection: collection,
	}
	reg := trace.StartRegion(ctx, name)
	span := &tracingSpan{
		openSpan:      zip,
		dataSegment:   seg,
		datastore:     true,
		runtimeRegion: reg,
	}
	if txnStarted {
		span.txn = txn
	}
	return span, ctx
}

func buildExternalSpan(ctx context.Context, name string, url string) (*tracingSpan, context.Context) {
	ctx, zip := ClientSpan(name, ctx)

	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if !strings.HasPrefix(url, "http") {
		url = "http://" + name + "/" + url
	}

	zip.SetTag("url", url)
	txnStarted := false
	txn := nrutil.GetNewRelicTransactionFromContext(ctx)
	if txn == nil {
		txnStarted = true
		ctx = nrutil.StartNRTransaction(name, ctx, nil, nil)
		txn = nrutil.GetNewRelicTransactionFromContext(ctx)
	}

	seg := newrelic.ExternalSegment{
		StartTime: txn.StartSegmentNow(),
		URL:       url,
	}
	reg := trace.StartRegion(ctx, name)
	span := &tracingSpan{
		openSpan:        zip,
		externalSegment: seg,
		external:        true,
		runtimeRegion:   reg,
	}
	if txnStarted {
		span.txn = txn
	}
	return span, ctx
}

// NewExternalSpan starts a span for tracing external actions
// This is used to trace actions against an external service, for example, a call to another service or a call to an external API
func NewExternalSpan(ctx context.Context, name string, url string) (Span, context.Context) {
	return buildExternalSpan(ctx, name, url)
}

// NewHTTPExternalSpan starts a span for tracing external HTTP actions
// This is used to trace actions against an external service, for example, a call to another service or a call to an external API
// It also adds the HTTP headers to the span so that the external service can trace the call back to this service if needed
func NewHTTPExternalSpan(ctx context.Context, name string, url string, hdr http.Header) (Span, context.Context) {
	s, ctx := buildExternalSpan(ctx, name, url)
	traceHTTPHeaders(ctx, s.openSpan, hdr)
	return s, ctx
}

func traceHTTPHeaders(ctx context.Context, sp opentracing.Span, hdr http.Header) {
	// Transmit the span's TraceContext as HTTP headers on our
	// outbound request.
	opentracing.GlobalTracer().Inject(
		sp.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(hdr))
}

// A type that conforms to opentracing.TextMapReader and
// opentracing.TextMapWriter.
type metadataReaderWriter struct {
	*metadata.MD
}

func (w metadataReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	if strings.HasSuffix(key, "-bin") {
		val = string(base64.StdEncoding.EncodeToString([]byte(val)))
	}
	(*w.MD)[key] = append((*w.MD)[key], val)
}

func (w metadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range *w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// ClientSpan starts a new client span linked to the existing spans if any are found
// in the context. The returned context should be used in place of the original
func ClientSpan(operationName string, ctx context.Context) (context.Context, opentracing.Span) {
	tracer := opentracing.GlobalTracer()
	var clientSpan opentracing.Span
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		clientSpan = tracer.StartSpan(
			operationName,
			opentracing.ChildOf(parentSpan.Context()),
		)
	} else {
		clientSpan = tracer.StartSpan(operationName)
	}
	otext.SpanKindRPCClient.Set(clientSpan)
	ctx = opentracing.ContextWithSpan(ctx, clientSpan)
	return ctx, clientSpan
}

// GRPCTracingSpan starts a new client span linked to the existing spans if any are found
// in the context. The returned context should be used in place of the original
func GRPCTracingSpan(operationName string, ctx context.Context) context.Context {
	tracer := opentracing.GlobalTracer()
	// Retrieve gRPC metadata.
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		md = md.Copy()
	} else {
		md = metadata.MD{}
	}
	if span := opentracing.SpanFromContext(ctx); span != nil {
		// There's nothing we can do with an error here.
		if err := tracer.Inject(span.Context(), opentracing.TextMap, metadataReaderWriter{&md}); err != nil {
			// log.Info(ctx, "err", err, "component", "tracing")
		}
	}

	var span opentracing.Span
	wireContext, err := tracer.Extract(opentracing.TextMap, metadataReaderWriter{&md})
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		// log.Info(ctx, "err", err, "component", "tracing")
	}
	span = tracer.StartSpan(operationName, otext.RPCServerOption(wireContext))
	ctx = opentracing.ContextWithSpan(ctx, span)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}
