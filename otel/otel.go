package otel

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

var tracer trace.Tracer

func SetTracer(tracerName string) {
	tracer = otel.Tracer(tracerName)
}

// StartSpan starts a new OTel span with the given name.
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if tracer == nil {
		return ctx, nil
	}
	return tracer.Start(ctx, spanName, opts...)
}

// helper to convert interface{} to string for setting otel attributes
func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return strings.TrimSpace(strings.ReplaceAll(fmt.Sprintf("%v", v), "\n", " "))
	}
}

// SetAttributes sets attributes on the given span.
func SetAttributes(span trace.Span, key string, value interface{}) {
	switch v := value.(type) {
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case string:
		span.SetAttributes(attribute.String(key, v))
	default:
		span.SetAttributes(attribute.String(key, toString(v)))
	}
}

// RecordError records an error on the span and sets status.
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

type metadataTextMapCarrier struct {
	md *metadata.MD
}

func (c metadataTextMapCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c metadataTextMapCarrier) Set(key, value string) {
	c.md.Set(key, value)
}

func (c metadataTextMapCarrier) Keys() []string {
	keys := make([]string, 0, len(*c.md))
	for k := range *c.md {
		keys = append(keys, k)
	}
	return keys
}

// InjectHTTPHeaders injects OTel context into HTTP headers.
func InjectHTTPHeaders(ctx context.Context, hdr http.Header) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(hdr))
}

// ExtractHTTPHeaders extracts OTel context from HTTP headers.
func ExtractHTTPHeaders(ctx context.Context, md metadata.MD) context.Context {
	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, &metadataTextMapCarrier{md: &md})
}
