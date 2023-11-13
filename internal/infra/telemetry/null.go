package telemetry

import (
	mnoop "go.opentelemetry.io/otel/metric/noop"
	tnoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
)

func ProvideNull(_ fx.Lifecycle) Telemetery {
	tel := Telemetery{
		serviceName:   "",
		namespace:     "",
		metricSrv:     nil,
		TraceProvider: tnoop.NewTracerProvider(),
		MeterProvider: mnoop.NewMeterProvider(),
	}

	return tel
}
