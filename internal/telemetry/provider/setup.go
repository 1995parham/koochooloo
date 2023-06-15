package provider

import (
	"fmt"
	"log"

	"github.com/1995parham/koochooloo/internal/telemetry/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
)

type Telemetery struct {
	serviceName   string
	namespace     string
	traceProvider *sdktrace.TracerProvider
	meterProvider *sdkmetric.MeterProvider
}

func New(cfg config.Config) Telemetery {
	var exporter sdktrace.SpanExporter

	var err error
	if !cfg.Trace.Enabled {
		exporter, err = stdout.New(
			stdout.WithPrettyPrint(),
		)
	} else {
		exporter, err = jaeger.New(
			jaeger.WithAgentEndpoint(jaeger.WithAgentHost(cfg.Trace.Agent.Host), jaeger.WithAgentPort(cfg.Trace.Agent.Port)),
		)
	}

	if err != nil {
		log.Fatalf("failed to initialize export pipeline for traces: %v", err)
	}

	reader, err := prometheus.New(prometheus.WithNamespace(cfg.Namespace))
	if err != nil {
		log.Fatalf("failed to initialize reader pipeline for metrics: %v", err)
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

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithResource(res), sdkmetric.WithReader(reader))

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	return Telemetery{
		serviceName:   cfg.ServiceName,
		namespace:     cfg.Namespace,
		traceProvider: tp,
		meterProvider: mp,
	}
}

func (t Telemetery) Meter() metric.Meter {
	return otel.Meter(fmt.Sprintf("%s/%s", t.namespace, t.serviceName))
}

func (t Telemetery) Trace() trace.Tracer {
	return otel.Tracer(fmt.Sprintf("%s/%s", t.namespace, t.serviceName))
}

func (t Telemetery) Shutdown(ctx context.Context) {
	if err := t.meterProvider.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}

	if err := t.traceProvider.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
