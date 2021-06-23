package trace

import (
	"log"

	"github.com/1995parham/koochooloo/internal/telemetry/config"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func New(cfg config.Trace) trace.Tracer {
	var exporter sdktrace.SpanExporter

	var err error
	if !cfg.Enabled {
		exporter, err = stdout.New(
			stdout.WithPrettyPrint(),
		)
	} else {
		exporter, err = jaeger.New(
			jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.URL + "/api/traces")))
	}

	if err != nil {
		log.Fatalf("failed to initialize export pipeline: %v", err)
	}


	res, err := resource.Merge(
			resource.Default(),
			resource.NewSchemaless(
				semconv.ServiceNamespaceKey.String("1995parham"),
				semconv.ServiceNameKey.String("koochooloo"),
			),
		)
		if err != nil {
			panic(err)
		}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(bsp), sdktrace.WithResource(res))

	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("1995parham.me/koochooloo")

	return tracer
}
