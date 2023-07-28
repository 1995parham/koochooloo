package telemetry

import (
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

func ProvideNull(_ fx.Lifecycle) Telemetery {
	tel := Telemetery{
		serviceName:   "",
		namespace:     "",
		metricSrv:     nil,
		TraceProvider: trace.NewNoopTracerProvider(),
		MeterProvider: noop.NewMeterProvider(),
	}

	return tel
}
