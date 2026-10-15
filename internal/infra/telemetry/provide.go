package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otelmetric "go.opentelemetry.io/otel/metric"
	mnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	tnoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
)

// setupTraceExporter returns nil when tracing is off and stdout export was not
// asked for: the caller then installs a no-op provider, so spans are never
// recorded. A disabled tracer used to print every span to stdout, which in a
// container meant a multi-line JSON block in the log for every request.
func setupTraceExporter(cfg Config) trace.SpanExporter {
	if !cfg.Trace.Enabled {
		if !cfg.Trace.Stdout {
			return nil
		}

		exporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			log.Fatalf("failed to initialize export pipeline for traces (stdout): %v", err)
		}

		return exporter
	}

	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(cfg.Trace.Endpoint), otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to initialize export pipeline for traces (otlp with grpc): %v", err)
	}

	return exporter
}

// setupMeterExporter returns a nil reader when metrics are off and stdout
// export was not asked for; the caller then installs a no-op provider.
func setupMeterExporter(cfg Config) (metric.Reader, *http.Server) {
	if !cfg.Meter.Enabled {
		if !cfg.Meter.Stdout {
			return nil, nil
		}

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
		// Zero selects http.DefaultMaxHeaderValueCount (Go 1.27).
		MaxHeaderValueCount: 0,
		TLSNextProto:        nil,
		ConnState:           nil,
		ErrorLog:            nil,
		BaseContext:         nil,
		ConnContext:         nil,
		HTTP2:               nil,
		Protocols:           nil,
		// False honours RFC 9218 client priority signals (Go 1.27 default).
		DisableClientPriority: false,
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

	var tp oteltrace.TracerProvider = tnoop.NewTracerProvider()

	if exporter != nil {
		bsp := trace.NewBatchSpanProcessor(exporter)
		tp = trace.NewTracerProvider(trace.WithSpanProcessor(bsp), trace.WithResource(res))
	}

	var mp otelmetric.MeterProvider = mnoop.NewMeterProvider()

	if reader != nil {
		mp = metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	}

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

func (t Telemetery) run(ctx context.Context) error {
	if t.metricSrv != nil {
		lc := new(net.ListenConfig)

		l, err := lc.Listen(ctx, "tcp", t.metricSrv.Addr)
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
	// metricSrv is nil when the meter is disabled (see setupMeterExporter),
	// mirroring the guard in run.
	if t.metricSrv == nil {
		return nil
	}

	if err := t.metricSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("cannot shutdown the metric server %w", err)
	}

	return nil
}
