// Package tracing provides distributed tracing for Go applications. It offers
// features such as collecting performance data, identifying where requests
// spend most of their time, and segmenting requests.
//
// Traces are created using OpenTracing APIs and exported via the configured
// global tracer (opentracing.GlobalTracer). The core package configures this
// tracer at startup — typically an OpenTelemetry bridge that sends traces to
// any OTLP-compatible backend (Jaeger, Grafana Tempo, Honeycomb, etc.) or
// New Relic.
package tracing

// SupportPackageIsVersion1 is a compile-time assertion constant.
// Downstream packages reference this to enforce version compatibility.
const SupportPackageIsVersion1 = true
