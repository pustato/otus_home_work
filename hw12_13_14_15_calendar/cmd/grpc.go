package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/app"
	grpcserver "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc"
	"github.com/spf13/cobra"
)

// grpcCmd represents the grpc command.
var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "Start grpc server",

	Run: func(cmd *cobra.Command, args []string) {
		configPath := cmd.Flag("config").Value.String()

		config := requireConfig(configPath)
		logg, cleanupLogger := requireLogger(config.Logger)
		defer cleanupLogger()

		eventRepo, cleanupEventRepo := requireEventStorage(config.Storage)
		defer cleanupEventRepo()

		events := app.NewEventUseCase(eventRepo)

		server := grpcserver.New(logg, events, config.GRPC.Addr())

		ctx, cancel := signal.NotifyContext(context.Background(),
			syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		defer cancel()

		go func() {
			<-ctx.Done()

			server.Stop()
		}()

		logg.Info("calendar is running over grpc...")

		if err := server.Start(); err != nil {
			logg.Error("failed to start grpc server: " + err.Error())
			cancel()
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(grpcCmd)
}
