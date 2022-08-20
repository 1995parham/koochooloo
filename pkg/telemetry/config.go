package telemetry

import (
	"github.com/1995parham/koochooloo/pkg/telemetry/log"
	"github.com/1995parham/koochooloo/pkg/telemetry/metric"
	"github.com/1995parham/koochooloo/pkg/telemetry/trace"
)

type Config struct {
	Log    *log.Config    `koanf:"log"`
	Metric *metric.Config `koanf:"metric"`
	Trace  *trace.Config  `koanf:"trace"`
}
