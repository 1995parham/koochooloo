package vectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type CounterVector struct {
	vector *prometheus.CounterVec
}

func NewCounter(namespace string, subsystem string, name string, labels []string) *CounterVector {
	vector := prometheus.NewCounterVec(prometheus.CounterOpts{
		Help:      fmt.Sprintf("counter vector for %s", name),
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
	}, labels)

	prometheus.MustRegister(vector)
	return &CounterVector{vector}
}

func (vector *CounterVector) Increment(labels ...string) {
	vector.vector.WithLabelValues(labels...).Inc()
}
