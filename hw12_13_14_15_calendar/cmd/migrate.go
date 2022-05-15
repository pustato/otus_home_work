package cmd

import (
	"database/sql"
	"embed"
	"fmt"
	"os"

	goose "github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var MigrationsFS embed.FS

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Set up migrations",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := cmd.Flag("config").Value.String()

		config := requireConfig(configPath)
		logg, cleanupLogger := requireLogger(config.Logger)
		defer cleanupLogger()

		if config.Storage.Driver != "db" {
			logg.Info("migrations are not required")
			os.Exit(0)
		}

		goose.SetBaseFS(MigrationsFS)

		db, err := sql.Open("pgx", config.Storage.dbConnectionString())
		if err != nil {
			logg.Error(fmt.Sprintf("cannot connect to DB: %v", err))
			resultCode = 1
			return
		}
		defer db.Close()

		if err := goose.SetDialect("postgres"); err != nil {
			logg.Error(fmt.Sprintf("migration prepare failed: %v", err))
			resultCode = 1
			return
		}

		if err := goose.Up(db, "migrations"); err != nil {
			logg.Error(fmt.Sprintf("migration up failed: %v", err))
			resultCode = 1
			return
		}

		logg.Info("migrations done")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
