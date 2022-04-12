package main

import (
	"embed"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/cmd"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	cmd.MigrationsFS = migrationsFS
	cmd.Execute()
}
