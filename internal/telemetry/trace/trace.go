package trace

import (
	"log"

	"github.com/1995parham/koochooloo/internal/telemetry/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

func New(cfg config.Trace) trace.Tracer {
	var exporter sdktrace.SpanExporter

	var err error
	if !cfg.Enabled {
		exporter, err = stdout.NewExporter(
			stdout.WithPrettyPrint(),
		)
	} else {
		exporter, err = jaeger.NewRawExporter(
			jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.URL + "/api/traces")))
	}

	if err != nil {
		log.Fatalf("failed to initialize export pipeline: %v", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(bsp), sdktrace.WithResource(
		resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.ServiceNamespaceKey.String("1995parham"),
				semconv.ServiceNameKey.String("koochooloo"),
			),
		),
	))

	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("1995parham.me/koochooloo")

	return tracer
}
