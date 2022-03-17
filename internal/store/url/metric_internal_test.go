package url

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewCounter(t *testing.T) {
	t.Parallel()

	c1 := register(prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "parham",
		Subsystem: "koochooloo",
		Name:      "insert_total",
		Help:      "total number of inserts",
		ConstLabels: map[string]string{
			"store": "url",
		},
	}))

	c2 := register(prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "parham",
		Subsystem: "koochooloo",
		Name:      "insert_total",
		Help:      "total number of inserts",
		ConstLabels: map[string]string{
			"store": "url",
		},
	}))

	require.Equal(t, c1, c2)
}
