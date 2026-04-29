package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattn/go-sqlite3"

	"project/storage"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}

	if path != ":memory:" {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Init(ctx context.Context) error {
	statements := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			link TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, link)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_links_user_created_at ON links(user_id, created_at)`,
		`CREATE TABLE IF NOT EXISTS message_counters (
			user_id INTEGER PRIMARY KEY,
			count INTEGER NOT NULL DEFAULT 0,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("init sqlite schema: %w", err)
		}
	}

	return nil
}

func (s *Storage) SaveLink(ctx context.Context, userID int64, link string) error {
	q := `INSERT INTO links (user_id, link) VALUES (?, ?)`

	if _, err := s.db.ExecContext(ctx, q, userID, link); err != nil {
		if isUniqueConstraint(err) {
			return storage.ErrDuplicateLink
		}

		return fmt.Errorf("save link: %w", err)
	}

	return nil
}

func (s *Storage) GetRandomLink(ctx context.Context, userID int64) (string, error) {
	q := `SELECT link FROM links WHERE user_id = ? ORDER BY RANDOM() LIMIT 1`

	var link string
	err := s.db.QueryRowContext(ctx, q, userID).Scan(&link)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrNoSavedPages
	}

	if err != nil {
		return "", fmt.Errorf("get random link: %w", err)
	}

	return link, nil
}

func (s *Storage) IncrementUserMessages(ctx context.Context, userID int64) (int, error) {
	q := `
		INSERT INTO message_counters (user_id, count, updated_at)
		VALUES (?, 1, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET
			count = count + 1,
			updated_at = CURRENT_TIMESTAMP
		RETURNING count
	`

	var count int
	if err := s.db.QueryRowContext(ctx, q, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("increment user messages: %w", err)
	}

	return count, nil
}

func (s *Storage) CountUserMessages(ctx context.Context, userID int64) (int, error) {
	q := `SELECT count FROM message_counters WHERE user_id = ?`

	var count int
	err := s.db.QueryRowContext(ctx, q, userID).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("count user messages: %w", err)
	}

	return count, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func isUniqueConstraint(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrConstraint
	}

	return false
}
