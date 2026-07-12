package urldb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// SQLURL stores and retrieves URLs in any GORM-supported SQL engine
// (sqlite, postgres, mysql). It implements urlrepo.Repository.
type SQLURL struct {
	DB      *gorm.DB
	Tracer  trace.Tracer
	Metrics Metrics
}

const one = 1

// ProvideDB creates a new URL store backed by the given GORM connection.
func ProvideDB(db *gorm.DB, tele telemetry.Telemetery) *SQLURL {
	tracer := tele.TraceProvider.Tracer("urldb.db")
	meter := tele.MeterProvider.Meter("urldb.db")

	return &SQLURL{
		DB:     db,
		Tracer: tracer,
		Metrics: Metrics{
			Usage:   NewUsage(meter, "sql"),
			Latency: NewLatency(meter),
		},
	}
}

// Migrate creates or updates the URL table for the configured dialect.
func Migrate(db *gorm.DB) error {
	//nolint:exhaustruct // AutoMigrate inspects the type only; the zero value is intentional.
	if err := db.AutoMigrate(&urlRecord{}); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	return nil
}

// liveExpr is the portable "not expired" predicate reused by the read paths.
// Expiry is enforced at query time (SQL has no TTL), mirroring the original
// MongoDB behaviour, and works identically on sqlite/postgres/mysql.
const liveExpr = "expire_time IS NULL OR expire_time >= ?"

// IncrementCount increments counter of url record by one, means url got visited.
func (s *SQLURL) IncrementCount(ctx context.Context, key string) error {
	ctx, span := s.Tracer.Start(ctx, "store.url.inc")
	defer span.End()

	// A single UPDATE ... SET count = count + 1 is atomic per row, so concurrent
	// increments stay consistent without an explicit transaction.
	rows, err := gorm.G[urlRecord](s.DB).
		Where("key = ?", key).
		Where(liveExpr, time.Now()).
		Update(ctx, "count", gorm.Expr("count + ?", one))
	if err != nil {
		span.RecordError(err)

		return fmt.Errorf("database failed: %w", err)
	}

	if rows == 0 {
		return urlrepo.ErrKeyNotFound
	}

	return nil
}

// Save saves given url in database.
func (s *SQLURL) Save(ctx context.Context, url model.URL) error {
	ctx, span := s.Tracer.Start(ctx, "store.url.save")
	defer span.End()

	start := time.Now()

	record := toRecord(url)
	if err := gorm.G[urlRecord](s.DB).Create(ctx, &record); err != nil {
		span.RecordError(err)

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return urlrepo.ErrDuplicateKey
		}

		return fmt.Errorf("database failed: %w", err)
	}

	s.Metrics.InsertedCounter.Add(ctx, 1)
	s.Metrics.InsertLatency.Record(ctx, time.Since(start).Seconds())

	return nil
}

// FindByKey retrieves url of the given key if it exists.
func (s *SQLURL) FindByKey(ctx context.Context, key string) (model.URL, error) {
	ctx, span := s.Tracer.Start(ctx, "store.url.find_by_key")
	defer span.End()

	record, err := gorm.G[urlRecord](s.DB).
		Where("key = ?", key).
		Where(liveExpr, time.Now()).
		First(ctx)
	if err != nil {
		span.RecordError(err)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.URL{}, urlrepo.ErrKeyNotFound
		}

		return model.URL{}, fmt.Errorf("database failed: %w", err)
	}

	s.Metrics.FetchedCounter.Add(ctx, 1)

	return toModel(record), nil
}
