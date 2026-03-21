// Package db provides embedded migrations and generated queries for the SQLite cache.
package db

import (
	"embed"
)

// Migrations FS.
//
//go:embed migrations/*.sql
var Migrations embed.FS
