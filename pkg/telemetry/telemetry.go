package telemetry

import (
	"github.com/1995parham/koochooloo/pkg/telemetry/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Telemetry struct {
	Log    *zap.Logger
	Metric metric.Metric
	Trace  trace.Tracer
}

func NewNoop() *Telemetry {
	return &Telemetry{
		Log:    zap.NewNop(),
		Metric: metric.New("namespace", "subsystem"),
		Trace:  trace.NewNoopTracerProvider().Tracer("namespace"),
	}
}
