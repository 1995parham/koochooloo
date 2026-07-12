package urlsvc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/generator"
	"github.com/1995parham/koochooloo/internal/domain/model"
	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
)

type URLSvc struct {
	repo urlrepo.Repository
	gen  generator.Generator
}

func Provide(repo urlrepo.Repository, gen generator.Generator) *URLSvc {
	return &URLSvc{
		gen:  gen,
		repo: repo,
	}
}

const maxRetries = 10

// Set creates an anonymous short URL (no owner), used by the public API.
func (s *URLSvc) Set(ctx context.Context, key, url string, expire *time.Time, count int) (string, error) {
	return s.set(ctx, key, url, expire, count, nil)
}

// SetForOwner creates a short URL owned by the given user, used by the admin
// panel. The visit count starts at zero.
func (s *URLSvc) SetForOwner(
	ctx context.Context, key, url string, expire *time.Time, ownerID uint,
) (string, error) {
	return s.set(ctx, key, url, expire, 0, &ownerID)
}

// ListByOwner returns every short URL owned by the given user.
func (s *URLSvc) ListByOwner(ctx context.Context, ownerID uint) ([]model.URL, error) {
	urls, err := s.repo.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("listing urls failed %w", err)
	}

	return urls, nil
}

// ListAll returns every short URL (admin view).
func (s *URLSvc) ListAll(ctx context.Context) ([]model.URL, error) {
	urls, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing urls failed %w", err)
	}

	return urls, nil
}

// Delete removes the short URL with the given key.
func (s *URLSvc) Delete(ctx context.Context, key string) error {
	if err := s.repo.Delete(ctx, key); err != nil {
		return fmt.Errorf("deleting url failed %w", err)
	}

	return nil
}

func (s *URLSvc) Get(ctx context.Context, key string) (model.URL, error) {
	u, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return model.URL{}, fmt.Errorf("database fetch failed %w", err)
	}

	return u, nil
}

// ResolveAndTrack retrieves the URL and increments its access count atomically.
func (s *URLSvc) ResolveAndTrack(ctx context.Context, key string) (model.URL, error) {
	u, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return model.URL{}, fmt.Errorf("database fetch failed %w", err)
	}

	if err := s.repo.IncrementCount(ctx, key); err != nil {
		return u, fmt.Errorf("database inc count failed %w", err)
	}

	return u, nil
}

func (s *URLSvc) Count(ctx context.Context, key string) (int, error) {
	u, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("database count failed %w", err)
	}

	return u.Count, nil
}

func (s *URLSvc) Inc(ctx context.Context, key string) error {
	if err := s.repo.IncrementCount(ctx, key); err != nil {
		return fmt.Errorf("database inc count failed %w", err)
	}

	return nil
}

// set saves a URL, honouring a custom key (prefixed with "$") or generating a
// unique one, and optionally attaching an owner.
func (s *URLSvc) set(
	ctx context.Context, key, url string, expire *time.Time, count int, owner *uint,
) (string, error) {
	if key != "" {
		key = "$" + key

		if err := s.repo.Save(ctx, model.URL{
			Key:        key,
			URL:        url,
			ExpireTime: expire,
			Count:      count,
			OwnerID:    owner,
		}); err != nil {
			if errors.Is(err, urlrepo.ErrDuplicateKey) {
				return "", fmt.Errorf("specified key is duplicated %w", err)
			}

			return "", fmt.Errorf("database insertion failed %w", err)
		}

		return key, nil
	}

	for range maxRetries {
		key = s.gen.ShortURLKey()

		if err := s.repo.Save(ctx, model.URL{
			Key:        key,
			URL:        url,
			ExpireTime: expire,
			Count:      count,
			OwnerID:    owner,
		}); err != nil {
			if errors.Is(err, urlrepo.ErrDuplicateKey) {
				continue
			}

			return "", fmt.Errorf("database insertion failed %w", err)
		}

		return key, nil
	}

	return "", fmt.Errorf("failed to generate unique key after %d attempts %w", maxRetries, urlrepo.ErrDuplicateKey)
}
