package provider

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/1995parham/koochooloo/internal/telemetry/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
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
	metricSrv     *http.ServeMux
	metricAddr    string
	traceProvider *sdktrace.TracerProvider
	meterProvider *sdkmetric.MeterProvider
}

func New(cfg config.Config) Telemetery {
	var exporter sdktrace.SpanExporter
	{
		var err error

		if !cfg.Trace.Enabled {
			exporter, err = stdouttrace.New(
				stdouttrace.WithPrettyPrint(),
			)
		} else {
			exporter, err = jaeger.New(
				jaeger.WithAgentEndpoint(jaeger.WithAgentHost(cfg.Trace.Agent.Host), jaeger.WithAgentPort(cfg.Trace.Agent.Port)),
			)
		}

		if err != nil {
			log.Fatalf("failed to initialize export pipeline for traces: %v", err)
		}
	}

	var reader sdkmetric.Reader
	var srv *http.ServeMux
	{
		var err error

		if !cfg.Meter.Enabled {
			exporter, err = stdoutmetric.New(
				stdoutmetric.WithPrettyPrint(),
			)
		} else {
			reader, err = prometheus.New(prometheus.WithNamespace(cfg.Namespace))

			srv = http.NewServeMux()
			srv.Handle("/metrics", promhttp.Handler())
		}

		if err != nil {
			log.Fatalf("failed to initialize reader pipeline for metrics: %v", err)
		}
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

func (t Telemetery) Meter() metric.Meter {
	return otel.Meter(fmt.Sprintf("%s/%s", t.namespace, t.serviceName))
}

func (t Telemetery) Trace() trace.Tracer {
	return otel.Tracer(fmt.Sprintf("%s/%s", t.namespace, t.serviceName))
}

func (t Telemetery) Shutdown(ctx context.Context) {
	if err := t.meterProvider.Shutdown(ctx); err != nil {
		log.Fatalf("cannot shutdown the meter provider: %v", err)
	}

	if err := t.traceProvider.Shutdown(ctx); err != nil {
		log.Fatalf("cannot shutdown the trace provider: %v", err)
	}
}
