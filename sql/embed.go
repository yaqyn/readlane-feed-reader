// Package sqlschema provides the migration files embedded in the executable.
package sqlschema

import "embed"

// Files contains the exact SQL used by Goose and sqlc.
//
//go:embed schema/*.sql
var Files embed.FS
