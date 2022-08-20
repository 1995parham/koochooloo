package vectors

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HistogramVector struct {
	vector *prometheus.HistogramVec
}

func NewHistogram(namespace string, subsystem string, name string, labels []string) *HistogramVector {
	vector := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Help:      fmt.Sprintf("histogram vector for %s", name),
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
	}, labels)

	prometheus.MustRegister(vector)
	return &HistogramVector{vector}
}

func (vector *HistogramVector) Observe(start time.Time, labels ...string) {
	vector.vector.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
}
