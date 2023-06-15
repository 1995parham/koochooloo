package url

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

// Usage contains metrics to meter database insert/retrieve.
type Usage struct {
	InsertedCounter metric.Int64Counter
	FetchedCounter  metric.Int64Counter
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
