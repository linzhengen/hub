package migrations

import (
	"embed"
)

//go:embed mysql/*.sql
var MySqlMigrationsFs embed.FS
