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
	// Verify the span can be ended without panic.
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
	// Set a query to exercise the datastore-specific path.
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

	// SetError with nil error should also be safe.
	err = span.SetError(nil)
	if err != nil {
		t.Fatal("expected SetError(nil) to return nil")
	}
}

func TestMetadataReaderWriter(t *testing.T) {
	md := metadata.MD{}
	rw := metadataReaderWriter{&md}

	// Test Set with a normal key.
	rw.Set("X-Request-ID", "abc123")
	vals := md["x-request-id"]
	if len(vals) != 1 || vals[0] != "abc123" {
		t.Fatalf("expected [abc123], got %v", vals)
	}

	// Test Set with a "-bin" suffix key triggers base64 encoding.
	rw.Set("X-Data-Bin", "hello")
	vals = md["x-data-bin"]
	if len(vals) != 1 || vals[0] != "aGVsbG8=" {
		t.Fatalf("expected [aGVsbG8=] for bin key, got %v", vals)
	}

	// Test that Set appends on repeated calls.
	rw.Set("X-Request-ID", "def456")
	vals = md["x-request-id"]
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d", len(vals))
	}

	// Test ForeachKey iterates all key-value pairs.
	collected := make(map[string][]string)
	err := rw.ForeachKey(func(key, val string) error {
		collected[key] = append(collected[key], val)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error from ForeachKey: %v", err)
	}
	if len(collected["x-request-id"]) != 2 {
		t.Fatalf("expected 2 values for x-request-id, got %d", len(collected["x-request-id"]))
	}
	if len(collected["x-data-bin"]) != 1 {
		t.Fatalf("expected 1 value for x-data-bin, got %d", len(collected["x-data-bin"]))
	}

	// Test ForeachKey propagates handler errors.
	expectedErr := errors.New("stop")
	err = rw.ForeachKey(func(key, val string) error {
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected ForeachKey to return handler error, got %v", err)
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
	span.Finish()

	// Test with an existing parent span in context.
	ctxWithSpan, parentSpan := ClientSpan("parent-operation", ctx)
	childCtx, childSpan := ClientSpan("child-of-parent", ctxWithSpan)
	if childSpan == nil {
		t.Fatal("expected non-nil child span")
	}
	if childCtx == nil {
		t.Fatal("expected non-nil child context")
	}
	childSpan.Finish()
	parentSpan.Finish()
}
