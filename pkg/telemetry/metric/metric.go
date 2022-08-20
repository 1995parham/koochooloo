package metric

import (
	"time"

	"github.com/1995parham/koochooloo/pkg/telemetry/metric/vectors"
)

type Metric interface {
	NewCounter(name string, labels []string) *vectors.CounterVector
	NewGauge(name string, labels []string) *vectors.GaugeVector
	NewHistogram(name string, labels []string) *vectors.HistogramVector
	NewSummary(name string, labels []string) *vectors.SummaryVector
}

type metric struct {
	Namespace string
	Subsystem string
}

func New(namespace string, subsystem string) Metric {
	return &metric{Namespace: namespace, Subsystem: subsystem}
}

// --------------------------------------------------- CounterVector

type CounterVector interface {
	Increment(labels ...string)
}

func (metric *metric) NewCounter(name string, labels []string) *vectors.CounterVector {
	return vectors.NewCounter(metric.Namespace, metric.Namespace, name, labels)
}

// --------------------------------------------------- GaugeVector

type GaugeVector interface {
	Increment(labels ...string)
	Decrement(labels ...string)
}

func (metric *metric) NewGauge(name string, labels []string) *vectors.GaugeVector {
	return vectors.NewGauge(metric.Namespace, metric.Namespace, name, labels)
}

// --------------------------------------------------- HistogramVector

type HistogramVector interface {
	Observe(start time.Time, labels ...string)
}

func (metric *metric) NewHistogram(name string, labels []string) *vectors.HistogramVector {
	return vectors.NewHistogram(metric.Namespace, metric.Namespace, name, labels)
}

// --------------------------------------------------- SummaryVector

type SummaryVector interface {
	Observe(start time.Time, labels ...string)
}

func (metric *metric) NewSummary(name string, labels []string) *vectors.SummaryVector {
	return vectors.NewSummary(metric.Namespace, metric.Namespace, name, labels)
}
