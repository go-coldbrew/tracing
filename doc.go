// Package tracing provides distributed tracing for Go applications. It offers
// features such as collecting performance data, identifying where requests
// spend most of their time, and segmenting requests.
//
// Traces are created and exported via OpenTelemetry. The core package
// configures the OTEL tracer provider at startup, sending traces to any
// OTLP-compatible backend (Jaeger, Grafana Tempo, Honeycomb, etc.) or
// New Relic.
package tracing

// SupportPackageIsVersion1 is a compile-time assertion constant.
// Downstream packages reference this to enforce version compatibility.
const SupportPackageIsVersion1 = true
