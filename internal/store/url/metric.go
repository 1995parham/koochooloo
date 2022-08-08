package url

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// Usage contains metrics to meter database insert/retrieve.
type Usage struct {
	InsertedCounter prometheus.Counter
	FetchedCounter  prometheus.Counter
}

// nolint: ireturn
func register[T prometheus.Collector](metric T) T {
	if err := prometheus.Register(metric); err != nil {
		var are prometheus.AlreadyRegisteredError
		if ok := errors.As(err, &are); ok {
			metric, ok = are.ExistingCollector.(T)
			if !ok {
				panic("different metric type registration")
			}
		} else {
			panic(err)
		}
	}

	return metric
}

func NewUsage(name string) Usage {
	inserted := register(prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "koochooloo",
		Name:      "inserted_total",
		Help:      "total number of insert operations",
		Subsystem: "url_store",
		ConstLabels: prometheus.Labels{
			"store": name,
		},
	}))

	fetched := register(prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "koochooloo",
		Name:      "fetched_total",
		Help:      "total number of fetch operations",
		Subsystem: "url_store",
		ConstLabels: prometheus.Labels{
			"store": name,
		},
	}))

	return Usage{
		InsertedCounter: inserted,
		FetchedCounter:  fetched,
	}
}
