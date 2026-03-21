package db

import (
	"embed"
)

// Migrations FS.
//
//go:embed migrations/*.sql
var Migrations embed.FS
