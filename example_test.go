package tracing_test

import (
	"context"

	"github.com/go-coldbrew/tracing"
)

func ExampleNewInternalSpan() {
	ctx := context.Background()

	// Create a span for internal business logic
	span, ctx := tracing.NewInternalSpan(ctx, "processOrder")
	defer span.End()

	// Your logic here — the span tracks duration and errors
	_ = ctx
}

func ExampleNewDatastoreSpan() {
	ctx := context.Background()

	// Create a span for a database operation
	span, ctx := tracing.NewDatastoreSpan(ctx, "postgres", "SELECT", "users")
	defer span.End()

	// Run your query — the span records datastore, operation, and collection
	_ = ctx
}

func ExampleNewExternalSpan() {
	ctx := context.Background()

	// Create a span for an external service call
	span, ctx := tracing.NewExternalSpan(ctx, "payment-gateway", "https://api.payments.example.com/charge")
	defer span.End()

	// Make the external call — the span tracks the dependency
	_ = ctx
}
