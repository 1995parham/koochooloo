package store

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// Usage contains metrics to meter database insert/retrieve.
type Usage struct {
	InsertedCounter prometheus.Counter
	FetchedCounter  prometheus.Counter
}

func NewUsage(name string) Usage {
	inserted := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "koochooloo",
		Name:      "inserted_total",
		Help:      "total number of insert operations",
		Subsystem: "url_store",
		ConstLabels: prometheus.Labels{
			"store": name,
		},
	})

	if err := prometheus.Register(inserted); err != nil {
		var are prometheus.AlreadyRegisteredError
		if ok := errors.As(err, &are); ok {
			inserted, ok = are.ExistingCollector.(prometheus.Counter)
			if !ok {
				panic("inserted must be a counter")
			}
		} else {
			panic(err)
		}
	}

	fetched := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "koochooloo",
		Name:      "fetched_total",
		Help:      "total number of fetch operations",
		Subsystem: "url_store",
		ConstLabels: prometheus.Labels{
			"store": name,
		},
	})

	if err := prometheus.Register(fetched); err != nil {
		var are prometheus.AlreadyRegisteredError
		if ok := errors.As(err, &are); ok {
			inserted, ok = are.ExistingCollector.(prometheus.Counter)
			if !ok {
				panic("fetched must be a counter")
			}
		} else {
			panic(err)
		}
	}

	return Usage{
		InsertedCounter: inserted,
		FetchedCounter:  fetched,
	}
}
