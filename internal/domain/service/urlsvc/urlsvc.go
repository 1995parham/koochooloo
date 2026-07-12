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

func (s *URLSvc) Set(ctx context.Context, key, url string, expire *time.Time, count int) (string, error) {
	if key != "" {
		key = "$" + key

		if err := s.repo.Save(ctx, model.URL{
			Key:        key,
			URL:        url,
			ExpireTime: expire,
			Count:      count,
			OwnerID:    nil,
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
			OwnerID:    nil,
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
