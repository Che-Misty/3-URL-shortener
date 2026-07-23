package url

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
)

type Storage interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error)
	DeleteURL(ctx context.Context, alias string) (int64, error)
	GetURL(ctx context.Context, alias string) (string, error)
	UpdateURL(ctx context.Context, alias string, newURL string) (int64, error)
	IsExist(ctx context.Context, alias string) (bool, error)
}

type UrlService struct {
	storage     Storage
	log         *slog.Logger
	aliasLength int
}

func New(storage Storage, log *slog.Logger, aliasLength int) *UrlService {
	return &UrlService{
		storage:     storage,
		log:         log,
		aliasLength: aliasLength,
	}
}

func (s *UrlService) SaveURL(ctx context.Context, urlToSave string, alias string) (string, error) {
	const op = "service.url.SaveURL"

	if alias == "" {
		for {
			alias = random.NewRandomString(s.aliasLength)

			exists, err := s.storage.IsExist(ctx, alias)
			if err != nil {
				return "", fmt.Errorf("%s: check alias existence: %w", op, err)
			}
			if !exists {
				break
			}

			s.log.Info("generated alias already exist", slog.String("alias", alias))
		}
	}

	id, err := s.storage.SaveURL(ctx, urlToSave, alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLExist) {
			s.log.Info("url already exist", slog.String("url", urlToSave))
			return "", fmt.Errorf("%s: %w", op, err)
		}
		return "", fmt.Errorf("%s: save url: %w", op, err)
	}

	s.log.Info("url added", slog.Int64("id", id))

	return alias, nil
}

func (s *UrlService) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "service.url.GetURL"

	resURL, err := s.storage.GetURL(ctx, alias)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			s.log.Info("url not found", slog.String("alias", alias))
			return "", fmt.Errorf("%s: %w", op, err)
		}
		s.log.Error("failed to get url", sl.Err(err))
		return "", fmt.Errorf("%s: get url: %w", op, err)
	}

	s.log.Info("got url", slog.String("url", resURL))

	return resURL, nil
}

func (s *UrlService) DeleteURL(ctx context.Context, alias string) (int64, error) {
	const op = "service.url.DeleteURL"

	countedDeleted, err := s.storage.DeleteURL(ctx, alias)
	if err != nil {
		s.log.Error("failed to delete url", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("url deleted", slog.Int64("countedDeleted", countedDeleted))

	return countedDeleted, nil
}

func (s *UrlService) UpdateURL(ctx context.Context, alias string, newURL string) (int64, error) {
	const op = "service.url.UpdateURL"

	rows, err := s.storage.UpdateURL(ctx, alias, newURL)
	if err != nil {
		s.log.Error("failed to update url", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("url updated", slog.String("alias", alias), slog.Int64("rows", rows))

	return rows, nil
}
