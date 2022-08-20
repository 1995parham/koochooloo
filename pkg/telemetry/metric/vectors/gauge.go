package vectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type GaugeVector struct {
	vector *prometheus.GaugeVec
}

func NewGauge(namespace string, subsystem string, name string, labels []string) *GaugeVector {
	vector := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Help:      fmt.Sprintf("gauge vector for %s", name),
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
	}, labels)

	prometheus.MustRegister(vector)
	return &GaugeVector{vector}
}

func (vector *GaugeVector) Increment(labels ...string) {
	vector.vector.WithLabelValues(labels...).Inc()
}

func (vector *GaugeVector) Decrement(labels ...string) {
	vector.vector.WithLabelValues(labels...).Dec()
}
