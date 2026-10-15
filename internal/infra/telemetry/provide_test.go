package telemetry_test

import (
	"context"
	"testing"

	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"
)

func config(stdout bool) telemetry.Config {
	return telemetry.Config{
		Trace:       telemetry.Trace{Enabled: false, Endpoint: "", Stdout: stdout},
		Meter:       telemetry.Meter{Address: "", Enabled: false, Stdout: stdout},
		Namespace:   "test",
		ServiceName: "koochooloo",
	}
}

func recording(t *testing.T, cfg telemetry.Config) bool {
	t.Helper()

	lc := fxtest.NewLifecycle(t)
	tel := telemetry.Provide(lc, cfg)
	lc.RequireStart()
	t.Cleanup(lc.RequireStop)

	_, span := tel.TraceProvider.Tracer("test").Start(context.Background(), "span")
	defer span.End()

	return span.IsRecording()
}

func TestDisabledTelemetryRecordsNothing(t *testing.T) {
	t.Parallel()

	require.False(t, recording(t, config(false)), "a disabled tracer must not record, or every span reaches stdout")
}

func TestDisabledTelemetryWithStdoutRecords(t *testing.T) {
	t.Parallel()

	require.True(t, recording(t, config(true)), "stdout = true keeps the old local-debugging export")
}
