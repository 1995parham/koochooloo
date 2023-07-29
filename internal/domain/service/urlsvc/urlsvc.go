package urlsvc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/1995parham/koochooloo/internal/domain/repository/urlrepo"
	"github.com/1995parham/koochooloo/internal/infra/generator"
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

func (s *URLSvc) Set(ctx context.Context, key, url string, expire *time.Time, count int) (string, error) {
	if key == "" {
		key = s.gen.ShortURLKey()
	} else {
		key = fmt.Sprintf("$%s", key)
	}

	if err := s.repo.Set(ctx, key, url, expire, count); err != nil {
		if errors.Is(err, urlrepo.ErrDuplicateKey) {
			if !strings.HasPrefix(key, "$") {
				// call set again to generate another random key.
				return s.Set(ctx, "", url, expire, 0)
			}

			return "", fmt.Errorf("specified key is duplicated %w", err)
		}

		return "", fmt.Errorf("database insertion failed %w", err)
	}

	return key, nil
}

func (s *URLSvc) Get(ctx context.Context, key string) (string, error) {
	url, err := s.repo.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("database fetch failed %w", err)
	}

	return url, nil
}

func (s *URLSvc) Count(ctx context.Context, key string) (int, error) {
	count, err := s.repo.Count(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("database count failed %w", err)
	}

	return count, nil
}

func (s *URLSvc) Inc(ctx context.Context, key string) error {
	if err := s.repo.Inc(ctx, key); err != nil {
		return fmt.Errorf("database inc count failed %w", err)
	}

	return nil
}
