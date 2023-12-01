package db

import (
	"database/sql"
)

// wrapped is a wrapper around a sql.DB with logging
type wrapped struct {
	*sql.DB
}

func newWrapped(db *sql.DB) *wrapped {
	return &wrapped{
		DB: db,
	}
}

func (w *wrapped) Exec(query string, args ...any) (sql.Result, error) {
	// slog.Debug("database exec", "q", query, "args", args)
	return w.DB.Exec(query, args...)
}

func (w *wrapped) Query(query string, args ...any) (*sql.Rows, error) {
	// slog.Debug("database query", "q", query, "args", args)
	return w.DB.Query(query, args...)
}

func (w *wrapped) QueryRow(query string, args ...any) *sql.Row {
	// slog.Debug("database row", "q", query, "args", args)
	return w.DB.QueryRow(query, args...)
}
