package vectors

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SummaryVector struct {
	vector *prometheus.SummaryVec
}

func NewSummary(namespace string, subsystem string, name string, labels []string) *SummaryVector {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Help:      fmt.Sprintf("summery vector for %s", name),
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
	}, labels)

	prometheus.MustRegister(vector)
	return &SummaryVector{vector}
}

func (vector *SummaryVector) Observe(start time.Time, labels ...string) {
	vector.vector.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
}
