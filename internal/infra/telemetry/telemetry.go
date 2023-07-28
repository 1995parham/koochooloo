package telemetry

import (
	"net/http"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Telemetery struct {
	serviceName   string
	namespace     string
	metricSrv     *http.Server
	TraceProvider trace.TracerProvider
	MeterProvider metric.MeterProvider
}
