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
func newCounter(counterOpts prometheus.CounterOpts) prometheus.Counter {
	ev := prometheus.NewCounter(counterOpts)

	if err := prometheus.Register(ev); err != nil {
		var are prometheus.AlreadyRegisteredError
		if ok := errors.As(err, &are); ok {
			ev, ok = are.ExistingCollector.(prometheus.Counter)
			if !ok {
				panic("different metric type registration")
			}
		} else {
			panic(err)
		}
	}

	return ev
}

func NewUsage(name string) Usage {
	inserted := newCounter(prometheus.CounterOpts{
		Namespace: "koochooloo",
		Name:      "inserted_total",
		Help:      "total number of insert operations",
		Subsystem: "url_store",
		ConstLabels: prometheus.Labels{
			"store": name,
		},
	})

	fetched := newCounter(prometheus.CounterOpts{
		Namespace: "koochooloo",
		Name:      "fetched_total",
		Help:      "total number of fetch operations",
		Subsystem: "url_store",
		ConstLabels: prometheus.Labels{
			"store": name,
		},
	})

	return Usage{
		InsertedCounter: inserted,
		FetchedCounter:  fetched,
	}
}
