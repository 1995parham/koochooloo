// Package apikeydb contains the SQL and in-memory implementations of the API
// key repository (GORM: sqlite, postgres, mysql).
package apikeydb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/apikeyrepo"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// SQLAPIKey stores and retrieves API keys in any GORM-supported SQL engine.
// It implements apikeyrepo.Repository.
type SQLAPIKey struct {
	DB     *gorm.DB
	Tracer trace.Tracer
}

// ProvideDB creates a new API key store backed by the given GORM connection.
func ProvideDB(db *gorm.DB, tele telemetry.Telemetery) *SQLAPIKey {
	return &SQLAPIKey{
		DB:     db,
		Tracer: tele.TraceProvider.Tracer("apikeydb.db"),
	}
}

// Migrate creates or updates the api_keys table for the configured dialect.
func Migrate(db *gorm.DB) error {
	//nolint:exhaustruct_v5 // AutoMigrate inspects the type only; the zero value is intentional.
	if err := db.AutoMigrate(&apiKeyRecord{}); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	return nil
}

// Create persists a new key and returns it with its assigned ID.
func (s *SQLAPIKey) Create(ctx context.Context, key model.APIKey) (model.APIKey, error) {
	ctx, span := s.Tracer.Start(ctx, "store.apikey.create")
	defer span.End()

	record := toRecord(key)
	if err := gorm.G[apiKeyRecord](s.DB).Create(ctx, &record); err != nil {
		span.RecordError(err)

		return model.APIKey{}, fmt.Errorf("database failed: %w", err)
	}

	return toModel(record), nil
}

// FindByHash returns the key whose SHA-256 hash is hash.
func (s *SQLAPIKey) FindByHash(ctx context.Context, hash string) (model.APIKey, error) {
	ctx, span := s.Tracer.Start(ctx, "store.apikey.find_by_hash")
	defer span.End()

	record, err := gorm.G[apiKeyRecord](s.DB).Where("hash = ?", hash).First(ctx)
	if err != nil {
		span.RecordError(err)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.APIKey{}, apikeyrepo.ErrAPIKeyNotFound
		}

		return model.APIKey{}, fmt.Errorf("database failed: %w", err)
	}

	return toModel(record), nil
}

// List returns all keys, ordered by id.
func (s *SQLAPIKey) List(ctx context.Context) ([]model.APIKey, error) {
	ctx, span := s.Tracer.Start(ctx, "store.apikey.list")
	defer span.End()

	records, err := gorm.G[apiKeyRecord](s.DB).Order("id").Find(ctx)
	if err != nil {
		span.RecordError(err)

		return nil, fmt.Errorf("database failed: %w", err)
	}

	return toModels(records), nil
}

// Revoke marks the key with the given id revoked at the given time.
func (s *SQLAPIKey) Revoke(ctx context.Context, id uint, at time.Time) error {
	ctx, span := s.Tracer.Start(ctx, "store.apikey.revoke")
	defer span.End()

	rows, err := gorm.G[apiKeyRecord](s.DB).Where("id = ?", id).Update(ctx, "revoked_at", at)
	if err != nil {
		span.RecordError(err)

		return fmt.Errorf("database failed: %w", err)
	}

	if rows == 0 {
		return apikeyrepo.ErrAPIKeyNotFound
	}

	return nil
}

// Touch records that the key with the given id was used at the given time.
func (s *SQLAPIKey) Touch(ctx context.Context, id uint, at time.Time) error {
	ctx, span := s.Tracer.Start(ctx, "store.apikey.touch")
	defer span.End()

	if _, err := gorm.G[apiKeyRecord](s.DB).Where("id = ?", id).Update(ctx, "last_used_at", at); err != nil {
		span.RecordError(err)

		return fmt.Errorf("database failed: %w", err)
	}

	return nil
}
