// Package userdb contains the SQL and in-memory implementations of the user
// repository (GORM: sqlite, postgres, mysql).
package userdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/userrepo"
	"github.com/1995parham/koochooloo/internal/infra/telemetry"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// SQLUser stores and retrieves users in any GORM-supported SQL engine.
// It implements userrepo.Repository.
type SQLUser struct {
	DB     *gorm.DB
	Tracer trace.Tracer
}

// ProvideDB creates a new user store backed by the given GORM connection.
func ProvideDB(db *gorm.DB, tele telemetry.Telemetery) *SQLUser {
	return &SQLUser{
		DB:     db,
		Tracer: tele.TraceProvider.Tracer("userdb.db"),
	}
}

// Migrate creates or updates the users table for the configured dialect.
func Migrate(db *gorm.DB) error {
	//nolint:exhaustruct // AutoMigrate inspects the type only; the zero value is intentional.
	if err := db.AutoMigrate(&userRecord{}); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	return nil
}

// Create persists a new user and returns it with its assigned ID.
func (s *SQLUser) Create(ctx context.Context, user model.User) (model.User, error) {
	ctx, span := s.Tracer.Start(ctx, "store.user.create")
	defer span.End()

	record := toRecord(user)
	if err := gorm.G[userRecord](s.DB).Create(ctx, &record); err != nil {
		span.RecordError(err)

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model.User{}, userrepo.ErrDuplicateUsername
		}

		return model.User{}, fmt.Errorf("database failed: %w", err)
	}

	return toModel(record), nil
}

// FindByUsername returns the user with the given username.
func (s *SQLUser) FindByUsername(ctx context.Context, username string) (model.User, error) {
	ctx, span := s.Tracer.Start(ctx, "store.user.find_by_username")
	defer span.End()

	record, err := gorm.G[userRecord](s.DB).Where("username = ?", username).First(ctx)

	return s.one(record, err, span)
}

// FindByID returns the user with the given id.
func (s *SQLUser) FindByID(ctx context.Context, id uint) (model.User, error) {
	ctx, span := s.Tracer.Start(ctx, "store.user.find_by_id")
	defer span.End()

	record, err := gorm.G[userRecord](s.DB).Where("id = ?", id).First(ctx)

	return s.one(record, err, span)
}

// FindBySubject returns the user federated from provider with the given subject.
func (s *SQLUser) FindBySubject(
	ctx context.Context, provider model.Provider, subject string,
) (model.User, error) {
	ctx, span := s.Tracer.Start(ctx, "store.user.find_by_subject")
	defer span.End()

	record, err := gorm.G[userRecord](s.DB).
		Where("provider = ? AND subject = ?", string(provider), subject).
		First(ctx)

	return s.one(record, err, span)
}

// List returns all users, ordered by id.
func (s *SQLUser) List(ctx context.Context) ([]model.User, error) {
	ctx, span := s.Tracer.Start(ctx, "store.user.list")
	defer span.End()

	records, err := gorm.G[userRecord](s.DB).Order("id").Find(ctx)
	if err != nil {
		span.RecordError(err)

		return nil, fmt.Errorf("database failed: %w", err)
	}

	return toModels(records), nil
}

// SetRole updates the role of the user with the given id.
func (s *SQLUser) SetRole(ctx context.Context, id uint, role model.Role) error {
	ctx, span := s.Tracer.Start(ctx, "store.user.set_role")
	defer span.End()

	rows, err := gorm.G[userRecord](s.DB).Where("id = ?", id).Update(ctx, "role", string(role))
	if err != nil {
		span.RecordError(err)

		return fmt.Errorf("database failed: %w", err)
	}

	if rows == 0 {
		return userrepo.ErrUserNotFound
	}

	return nil
}

// Delete removes the user with the given id.
func (s *SQLUser) Delete(ctx context.Context, id uint) error {
	ctx, span := s.Tracer.Start(ctx, "store.user.delete")
	defer span.End()

	rows, err := gorm.G[userRecord](s.DB).Where("id = ?", id).Delete(ctx)
	if err != nil {
		span.RecordError(err)

		return fmt.Errorf("database failed: %w", err)
	}

	if rows == 0 {
		return userrepo.ErrUserNotFound
	}

	return nil
}

// one maps a single-row query result onto the domain model, translating
// gorm.ErrRecordNotFound into userrepo.ErrUserNotFound.
func (s *SQLUser) one(record userRecord, err error, span trace.Span) (model.User, error) {
	if err != nil {
		span.RecordError(err)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, userrepo.ErrUserNotFound
		}

		return model.User{}, fmt.Errorf("database failed: %w", err)
	}

	return toModel(record), nil
}
