package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	app "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/application"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/server/httpserver"
	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Start http server",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := cmd.Flag("config").Value.String()

		config := requireConfig(configPath)
		logg, cleanupLogger := requireLogger(config.Logger)
		defer cleanupLogger()

		eventRepo, cleanupEventRepo := requireEventRepo(config.Storage)
		defer cleanupEventRepo()

		events := app.NewEventUseCase(eventRepo)

		server := httpserver.New(logg, events, config.HTTP.Addr())

		ctx, cancel := signal.NotifyContext(context.Background(),
			syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		defer cancel()

		go func() {
			<-ctx.Done()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			if err := server.Stop(ctx); err != nil {
				logg.Error("failed to stop http server: " + err.Error())
			}
		}()

		logg.Info("calendar is running...")

		if err := server.Start(ctx); err != nil {
			logg.Error("failed to start http server: " + err.Error())
			cancel()
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)
}
