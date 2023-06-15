package telemetry

import (
	"errors"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"golang.org/x/net/context"
)

type Telemetery struct {
	serviceName   string
	namespace     string
	metricSrv     *http.ServeMux
	metricAddr    string
	traceProvider *trace.TracerProvider
	meterProvider *metric.MeterProvider
}

func setupTraceExporter(cfg Config) trace.SpanExporter {
	if !cfg.Trace.Enabled {
		exporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			log.Fatalf("failed to initialize export pipeline for traces (stdout): %v", err)
		}

		return exporter
	}

	exporter, err := jaeger.New(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(cfg.Trace.Agent.Host),
			jaeger.WithAgentPort(cfg.Trace.Agent.Port),
		),
	)
	if err != nil {
		log.Fatalf("failed to initialize export pipeline for traces (jeager): %v", err)
	}

	return exporter
}

func setupMeterExporter(cfg Config) (metric.Reader, *http.ServeMux) {
	if !cfg.Meter.Enabled {
		exporter, err := stdoutmetric.New()
		if err != nil {
			log.Fatalf("failed to initialize reader pipeline for metrics (stdout): %v", err)
		}

		return metric.NewPeriodicReader(exporter), nil
	}

	exporter, err := prometheus.New(prometheus.WithNamespace(cfg.Namespace))
	if err != nil {
		log.Fatalf("failed to initialize reader pipeline for metrics (prometheus): %v", err)
	}

	srv := http.NewServeMux()
	srv.Handle("/metrics", promhttp.Handler())

	return exporter, srv
}

func New(cfg Config) Telemetery {
	reader, srv := setupMeterExporter(cfg)
	exporter := setupTraceExporter(cfg)

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

	bsp := trace.NewBatchSpanProcessor(exporter)
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(bsp), trace.WithResource(res))

	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	return Telemetery{
		serviceName:   cfg.ServiceName,
		namespace:     cfg.Namespace,
		metricSrv:     srv,
		metricAddr:    cfg.Meter.Address,
		traceProvider: tp,
		meterProvider: mp,
	}
}

func (t Telemetery) Run() {
	if t.metricSrv != nil {
		go func() {
			// nolint: gosec
			if err := http.ListenAndServe(t.metricAddr, t.metricSrv); !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("metric server initiation failed: %v", err)
			}
		}()
	}
}

func (t Telemetery) Shutdown(ctx context.Context) {
	if err := t.meterProvider.Shutdown(ctx); err != nil {
		log.Fatalf("cannot shutdown the meter provider: %v", err)
	}

	if err := t.traceProvider.Shutdown(ctx); err != nil {
		log.Fatalf("cannot shutdown the trace provider: %v", err)
	}
}
