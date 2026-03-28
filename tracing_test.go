package tracing

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestNewInternalSpan(t *testing.T) {
	ctx := context.Background()
	span, newCtx := NewInternalSpan(ctx, "test-internal")
	if span == nil {
		t.Fatal("expected non-nil span")
	}
	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}
	// With the noop tracer, the span context won't be valid,
	// but we verify a span exists and can be ended without panic.
	span.End()
}

func TestNewDatastoreSpan(t *testing.T) {
	ctx := context.Background()
	span, newCtx := NewDatastoreSpan(ctx, "redis", "GET", "users")
	if span == nil {
		t.Fatal("expected non-nil span")
	}
	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}
	span.SetQuery("GET users:123")
	span.SetTag("key", "users:123")
	span.End()
}

func TestNewExternalSpan(t *testing.T) {
	ctx := context.Background()
	span, newCtx := NewExternalSpan(ctx, "other-service", "/api/v1/resource")
	if span == nil {
		t.Fatal("expected non-nil span")
	}
	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}
	span.SetTag("status", 200)
	span.End()
}

func TestSpanNilSafety(t *testing.T) {
	var span *tracingSpan

	// All methods on a nil *tracingSpan should be safe to call.
	span.End()
	span.Finish()
	span.SetTag("key", "value")
	span.SetQuery("SELECT 1")

	err := span.SetError(errors.New("some error"))
	if err == nil {
		t.Fatal("expected SetError to return the provided error even on nil span")
	}

	err = span.SetError(nil)
	if err != nil {
		t.Fatal("expected SetError(nil) to return nil")
	}
}

func TestMetadataCarrier(t *testing.T) {
	md := metadata.MD{}
	mc := metadataCarrier(md)

	mc.Set("x-request-id", "abc123")
	if got := mc.Get("x-request-id"); got != "abc123" {
		t.Fatalf("expected abc123, got %s", got)
	}

	if got := mc.Get("nonexistent"); got != "" {
		t.Fatalf("expected empty string for missing key, got %s", got)
	}

	keys := mc.Keys()
	if len(keys) != 1 || keys[0] != "x-request-id" {
		t.Fatalf("expected [x-request-id], got %v", keys)
	}
}

func TestClientSpan(t *testing.T) {
	ctx := context.Background()
	newCtx, span := ClientSpan("child-operation", ctx)
	if span == nil {
		t.Fatal("expected non-nil span from ClientSpan")
	}
	if newCtx == nil {
		t.Fatal("expected non-nil context from ClientSpan")
	}
	span.End()

	// Test with an existing parent span in context.
	ctxWithSpan, parentSpan := ClientSpan("parent-operation", ctx)
	childCtx, childSpan := ClientSpan("child-of-parent", ctxWithSpan)
	if childSpan == nil {
		t.Fatal("expected non-nil child span")
	}
	if childCtx == nil {
		t.Fatal("expected non-nil child context")
	}
	childSpan.End()
	parentSpan.End()
}

func TestGRPCTracingSpan(t *testing.T) {
	ctx := context.Background()
	newCtx := GRPCTracingSpan("test-operation", ctx)
	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}

	// Test with existing span in context (via ClientSpan).
	ctxWithSpan, span := ClientSpan("parent-op", ctx)
	defer span.End()
	newCtx2 := GRPCTracingSpan("child-operation", ctxWithSpan)
	if newCtx2 == nil {
		t.Fatal("expected non-nil context with parent span")
	}
}

func TestNewHTTPExternalSpan(t *testing.T) {
	ctx := context.Background()
	hdr := make(map[string][]string)
	span, newCtx := NewHTTPExternalSpan(ctx, "external-svc", "/api/data", hdr)
	if span == nil {
		t.Fatal("expected non-nil span")
	}
	if newCtx == nil {
		t.Fatal("expected non-nil context")
	}
	// Headers may contain trace propagation if a real propagator is configured.
	span.End()
}
