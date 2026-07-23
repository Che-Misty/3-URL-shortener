package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"url-shortener/internal/storage"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Storage struct {
	db *pgxpool.Pool
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	if err := runMigrations(storagePath); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	db, err := pgxpool.New(context.Background(), storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func runMigrations(storagePath string) error {
	const op = "storage.postgres.runMigrations"

	db, err := sql.Open("pgx", storagePath)
	if err != nil {
		return fmt.Errorf("%s: open db: %w", op, err)
	}
	defer db.Close()

	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("%s: init driver: %w", op, err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("%s: init source: %w", op, err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "pgx5", driver)
	if err != nil {
		return fmt.Errorf("%s: init migrate: %w", op, err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("%s: apply migrations: %w", op, err)
	}

	return nil
}

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	var id int64
	err := s.db.QueryRow(ctx,
		`
		INSERT INTO url (alias, url)
		VALUES ($1, $2)
		RETURNING id
		`,
		alias,
		urlToSave,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, storage.ErrURLExist
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	var resUrl string

	err := s.db.QueryRow(ctx,
		"SELECT url FROM url WHERE alias = $1", alias).Scan(&resUrl)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resUrl, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) (int64, error) {
	const op = "storage.postgres.DeleteURL"

	result, err := s.db.Exec(ctx,
		"DELETE FROM url WHERE alias = $1", alias)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return result.RowsAffected(), nil
}

func (s *Storage) UpdateURL(ctx context.Context, alias string, newURL string) (int64, error) {
	const op = "storage.postgres.UpdateURL"

	result, err := s.db.Exec(ctx,
		"UPDATE url SET url = $1 WHERE alias = $2", newURL, alias)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return result.RowsAffected(), nil
}

func (s *Storage) IsExist(ctx context.Context, alias string) (bool, error) {
	const op = "storage.postgres.IsExist"

	var exists bool
	err := s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM url WHERE alias = $1)", alias).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

func (s *Storage) Close() error {
	s.db.Close()
	return nil
}
