// Package migrations applies versioned, embedded PostgreSQL migrations with Goose.
package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
	sqlschema "github.com/yaqyn/readlane/sql"
	"io/fs"
	"log/slog"
)

func Provider(db *sql.DB) (*goose.Provider, error) {
	files, err := fs.Sub(sqlschema.Files, "schema")
	if err != nil {
		return nil, err
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, err
	}
	return goose.NewProvider(goose.DialectPostgres, db, files, goose.WithSessionLocker(locker), goose.WithSlog(slog.Default()))
}
func Up(ctx context.Context, db *sql.DB) error {
	p, err := Provider(db)
	if err != nil {
		return err
	}
	_, err = p.Up(ctx)
	return err
}
