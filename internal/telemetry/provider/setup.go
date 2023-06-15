package provider

import (
	"fmt"
	"log"

	"github.com/1995parham/koochooloo/internal/telemetry/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func New(cfg config.Trace) (trace.Tracer, metric.Meter) {
	var exporter sdktrace.SpanExporter

	var err error
	if !cfg.Enabled {
		exporter, err = stdout.New(
			stdout.WithPrettyPrint(),
		)
	} else {
		exporter, err = jaeger.New(
			jaeger.WithAgentEndpoint(jaeger.WithAgentHost(cfg.Agent.Host), jaeger.WithAgentPort(cfg.Agent.Port)),
		)
	}

	if err != nil {
		log.Fatalf("failed to initialize export pipeline: %v", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNamespaceKey.String(cfg.Namespace),
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		panic(err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(bsp), sdktrace.WithResource(res))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithResource(res))

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	tracer := otel.Tracer(fmt.Sprintf("%s/%s", cfg.Namespace, cfg.ServiceName))
	meter := otel.Meter(fmt.Sprintf("%s/%s", cfg.Namespace, cfg.ServiceName))

	return tracer, meter
}
