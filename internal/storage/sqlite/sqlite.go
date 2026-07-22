package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {

	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS shortener (
			id INTEGER PRIMARY KEY,
			alias TEXT UNIQUE NOT NULL,
			url TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
		CREATE INDEX IF NOT EXISTS idx_alias ON shortener(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	//INSERT OR REPLACE INTO shortener (alias, url) VALUES (?, ?)
	stmt, err := s.db.Prepare(`
		INSERT INTO shortener (alias, url) VALUES (?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	//defer stmt.Close()

	result, err := stmt.Exec(alias, urlToSave)
	if err != nil {
		//рефактор ошибки дубликатов
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExist)
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare(`
	SELECT url FROM shortener WHERE alias = ?
	`)
	if err != nil {
		return "", fmt.Errorf("%s: %v", op, err)
	}

	var resUrl string

	err = stmt.QueryRow(alias).Scan(&resUrl)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %v", op, err)
	}

	return resUrl, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare(`
	DELETE FROM shortener WHERE alias = ?
	`)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	_, err = stmt.Exec(alias)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrURLNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil
}
