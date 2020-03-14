package store

import "github.com/prometheus/client_golang/prometheus"

// Usage contains metrics to meter database insert/retrieve
type Usage struct {
	InsertedCounter prometheus.Counter
	FetchedCounter  prometheus.Counter
}

func NewUsage(name string) Usage {
	inserted := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "koochooloo",
		Name:      "inserted_counter",
		ConstLabels: prometheus.Labels{
			"store": name,
		},
	})

	if err := prometheus.Register(inserted); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			inserted = are.ExistingCollector.(prometheus.Counter)
		} else {
			panic(err)
		}
	}

	fetched := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "koochooloo",
		Name:      "fetched_counter",
		ConstLabels: prometheus.Labels{
			"store": name,
		},
	})

	if err := prometheus.Register(fetched); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			inserted = are.ExistingCollector.(prometheus.Counter)
		} else {
			panic(err)
		}
	}

	return Usage{
		InsertedCounter: inserted,
		FetchedCounter:  fetched,
	}
}
