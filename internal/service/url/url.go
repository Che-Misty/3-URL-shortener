package url

import (
	"errors"
	"fmt"
	"log/slog"

	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
)

type Storage interface {
	SaveURL(urlToSave string, alias string) (int64, error)
	DeleteURL(alias string) (int64, error)
	GetURL(alias string) (string, error)
	UpdateURL(alias string, newURL string) (int64, error)
	IsExist(alias string) (exists bool)
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

func (s *UrlService) SaveURL(urlToSave string, alias string) (string, error) {
	const op = "service.url.SaveURL"

	if alias == "" {
		for {
			alias = random.NewRandomString(s.aliasLength)
			if !s.storage.IsExist(alias) {
				break
			}
			s.log.Info("generated alias already exist", slog.String("alias", alias))

		}
	}

	id, err := s.storage.SaveURL(urlToSave, alias)
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

func (s *UrlService) GetURL(alias string) (string, error) {
	const op = "service.url.GetURL"

	resURL, err := s.storage.GetURL(alias)
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

func (s *UrlService) DeleteURL(alias string) (int64, error) {
	const op = "service.url.DeleteURL"

	countedDeleted, err := s.storage.DeleteURL(alias)
	if err != nil {
		s.log.Error("failed to delete url", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("url deleted", slog.Int("countedDeleted", int(countedDeleted)))

	return countedDeleted, nil
}

func (s *UrlService) UpdateURL(alias string, newURL string) (int64, error) {
	const op = "service.url.UpdateURL"

	rows, err := s.storage.UpdateURL(alias, newURL)
	if err != nil {
		s.log.Error("failed to update url", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}
	s.log.Info("url updated", slog.String("alias", alias), slog.Int64("rows", int64(rows)))
	return rows, nil
}
