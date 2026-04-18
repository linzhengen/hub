package migrations

import (
	"embed"
)

//go:embed postgres/*.sql
var PostgresMigrationsFs embed.FS
