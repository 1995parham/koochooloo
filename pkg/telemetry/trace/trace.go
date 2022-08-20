package trace

import (
	"fmt"
	"log"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func New(cfg *Config, namespace string, subsystem string) trace.Tracer {
	instrumentation := fmt.Sprintf("%s/%s", namespace, subsystem)

	if !cfg.Enabled {
		return trace.NewNoopTracerProvider().Tracer(instrumentation)
	}

	exporter, err := jaeger.New(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(cfg.Host),
			jaeger.WithAgentPort(strconv.Itoa(cfg.Port)),
		),
	)
	if err != nil {
		log.Fatalf("failed to initialize export pipeline: %v", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNamespaceKey.String(namespace),
			semconv.ServiceNameKey.String(subsystem),
		),
	)
	if err != nil {
		panic(err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Ratio))),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	// register the TraceContext propagator globally.
	var tc propagation.TraceContext

	otel.SetTextMapPropagator(tc)

	tracer := otel.Tracer(instrumentation)

	return tracer
}
