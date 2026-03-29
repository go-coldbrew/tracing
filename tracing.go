package tracing

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"runtime/trace"
	"strings"

	nrutil "github.com/go-coldbrew/tracing/newrelic"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

const tracerName = "github.com/go-coldbrew/tracing"

// toAttribute converts a key-value pair to a typed OTEL attribute,
// preserving numeric and boolean types instead of stringifying everything.
func toAttribute(key string, value any) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	default:
		return attribute.String(key, fmt.Sprint(v))
	}
}

// Span defines an interface for implementing a tracing span.
// Consumers use this to create and annotate spans without coupling to a
// specific tracing backend.
type Span interface {
	// End ends the span, can also use Finish()
	End()
	// Finish ends the span, can also use End()
	Finish()
	// SetTag sets a tag on the span, can be used to add custom attributes
	SetTag(key string, value any)
	// SetQuery sets the query on the span, can be used to add query for datastore spans
	SetQuery(query string)
	// SetError sets the error on the span
	SetError(err error) error
}

type tracingSpan struct {
	otelSpan        oteltrace.Span
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
		return
	}
	span.otelSpan.End()

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

func (span *tracingSpan) SetTag(key string, value any) {
	if span == nil {
		return
	}
	span.otelSpan.SetAttributes(toAttribute(key, value))
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
		return
	}
	if span.datastore {
		span.otelSpan.SetAttributes(semconv.DBQueryText(query))
		span.dataSegment.ParameterizedQuery = query
	} else {
		span.otelSpan.SetAttributes(attribute.String("query", query))
	}
}

func (span *tracingSpan) SetError(err error) error {
	if span == nil || err == nil {
		return err
	}
	span.otelSpan.SetStatus(codes.Error, err.Error())
	span.otelSpan.RecordError(err)
	if span.datastore {
		span.dataSegment.AddAttribute("error", err)
	} else if span.external {
		span.externalSegment.AddAttribute("error", err)
	} else {
		span.segment.AddAttribute("error", err)
	}

	if span.txn != nil {
		span.txn.NoticeError(err)
	}
	return err
}

// NewInternalSpan starts a span for tracing internal actions.
// This is used to trace actions within the same service, for example, a function call.
func NewInternalSpan(ctx context.Context, name string) (Span, context.Context) {
	ctx, otelSpan := otel.Tracer(tracerName).Start(ctx, name)

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
		otelSpan:      otelSpan,
		segment:       seg,
		runtimeRegion: reg,
	}
	if txnStarted {
		span.txn = txn
	}
	return span, ctx
}

// NewDatastoreSpan starts a span for tracing data store actions.
// This is used to trace actions against a data store, for example, a database query or a redis call.
func NewDatastoreSpan(ctx context.Context, datastore, operation, collection string) (Span, context.Context) {
	name := operation
	if !strings.HasPrefix(name, datastore) {
		name = datastore + name
	}
	ctx, otelSpan := otel.Tracer(tracerName).Start(ctx, name,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
	)
	otelSpan.SetAttributes(
		semconv.DBSystemNameKey.String(datastore),
		semconv.DBCollectionName(collection),
		semconv.DBOperationName(operation),
	)

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
		otelSpan:      otelSpan,
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
	ctx, clientSpan := clientSpanOTEL(ctx, name)

	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if !strings.HasPrefix(url, "http") {
		url = "http://" + name + "/" + url
	}

	serverAddr := name
	if parsed, err := neturl.Parse(url); err == nil && parsed.Hostname() != "" {
		serverAddr = parsed.Hostname()
	}
	clientSpan.SetAttributes(
		semconv.URLFull(url),
		semconv.ServerAddress(serverAddr),
	)
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
		otelSpan:        clientSpan,
		externalSegment: seg,
		external:        true,
		runtimeRegion:   reg,
	}
	if txnStarted {
		span.txn = txn
	}
	return span, ctx
}

// NewExternalSpan starts a span for tracing external actions.
// This is used to trace actions against an external service.
func NewExternalSpan(ctx context.Context, name string, url string) (Span, context.Context) {
	return buildExternalSpan(ctx, name, url)
}

// NewHTTPExternalSpan starts a span for tracing external HTTP actions.
// It also injects trace propagation headers so the external service can
// correlate the call back to this service.
func NewHTTPExternalSpan(ctx context.Context, name string, url string, hdr http.Header) (Span, context.Context) {
	s, ctx := buildExternalSpan(ctx, name, url)
	if hdr != nil {
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(hdr))
	}
	return s, ctx
}

// clientSpanOTEL starts a new client span linked to any existing span in context.
func clientSpanOTEL(ctx context.Context, operationName string) (context.Context, oteltrace.Span) {
	return otel.Tracer(tracerName).Start(ctx, operationName,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
	)
}

// ClientSpan starts a new client span linked to the existing spans if any are found
// in the context. The returned context should be used in place of the original.
func ClientSpan(operationName string, ctx context.Context) (context.Context, oteltrace.Span) {
	return clientSpanOTEL(ctx, operationName)
}

// GRPCTracingSpan starts a new server span from incoming gRPC metadata.
// The returned context should be used in place of the original.
func GRPCTracingSpan(operationName string, ctx context.Context) context.Context {
	// Extract trace context from incoming gRPC metadata.
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		md = md.Copy()
	} else {
		md = metadata.MD{}
	}

	// Extract propagated trace context from metadata.
	prop := otel.GetTextMapPropagator()
	ctx = prop.Extract(ctx, metadataCarrier(md))

	// Start a server span (automatically linked to extracted parent).
	ctx, _ = otel.Tracer(tracerName).Start(ctx, operationName,
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	)

	// Preserve existing outgoing metadata and inject trace context into it.
	outMD, _ := metadata.FromOutgoingContext(ctx)
	outMD = outMD.Copy()
	prop.Inject(ctx, metadataCarrier(outMD))
	ctx = metadata.NewOutgoingContext(ctx, outMD)
	return ctx
}

// metadataCarrier adapts gRPC metadata.MD to propagation.TextMapCarrier.
type metadataCarrier metadata.MD

func (mc metadataCarrier) Get(key string) string {
	vals := metadata.MD(mc).Get(key)
	if len(vals) == 0 {
		return ""
	}
	// Join multiple values for W3C baggage/tracestate compatibility.
	return strings.Join(vals, ",")
}

func (mc metadataCarrier) Set(key, value string) {
	metadata.MD(mc).Set(key, value)
}

func (mc metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for k := range mc {
		keys = append(keys, k)
	}
	return keys
}
