package telemetry

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

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
	"go.uber.org/fx"
	"golang.org/x/net/context"
)

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

func setupMeterExporter(cfg Config) (metric.Reader, *http.Server) {
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

	return exporter, &http.Server{
		Addr:                         cfg.Meter.Address,
		Handler:                      srv,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  time.Second,
		ReadHeaderTimeout:            time.Second,
		WriteTimeout:                 time.Second,
		IdleTimeout:                  time.Second,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
	}
}

func Provide(lc fx.Lifecycle, cfg Config) Telemetery {
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

	tel := Telemetery{
		serviceName:   cfg.ServiceName,
		namespace:     cfg.Namespace,
		metricSrv:     srv,
		TraceProvider: tp,
		MeterProvider: mp,
	}

	lc.Append(
		fx.Hook{
			OnStart: tel.run,
			OnStop:  tel.shutdown,
		},
	)

	return tel
}

func (t Telemetery) run(_ context.Context) error {
	if t.metricSrv != nil {
		l, err := net.Listen("tcp", t.metricSrv.Addr)
		if err != nil {
			return fmt.Errorf("metric server listen failed: %w", err)
		}

		go func() {
			if err := t.metricSrv.Serve(l); !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("metric server initiation failed: %v", err)
			}
		}()
	}

	return nil
}

func (t Telemetery) shutdown(ctx context.Context) error {
	if err := t.metricSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("cannot shutdown the metric server %w", err)
	}

	return nil
}
