package urldb

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	Latency
	Usage
}

// Latency contains metrics to meter database latency on insert and retrieve.
type Latency struct {
	InsertLatency metric.Float64Histogram
}

// Usage contains metrics to meter database insert/retrieve.
type Usage struct {
	InsertedCounter metric.Int64Counter
	FetchedCounter  metric.Int64Counter
}

func NewLatency(meter metric.Meter) Latency {
	insert, err := meter.Float64Histogram(
		"store.url.insert.latency",
		metric.WithUnit("s"),
		metric.WithDescription("latency of database inserts"),
	)
	if err != nil {
		panic(err)
	}

	return Latency{
		InsertLatency: insert,
	}
}

func NewUsage(meter metric.Meter, name string) Usage {
	inserted, err := meter.Int64Counter(
		fmt.Sprintf("store.url.%s.inserted", name),
		metric.WithDescription("total number of insert operations"),
	)
	if err != nil {
		panic(err)
	}

	fetched, err := meter.Int64Counter(
		fmt.Sprintf("store.url.%s.fetched", name),
		metric.WithDescription("total number of fetch operations"),
	)
	if err != nil {
		panic(err)
	}

	return Usage{
		InsertedCounter: inserted,
		FetchedCounter:  fetched,
	}
}
