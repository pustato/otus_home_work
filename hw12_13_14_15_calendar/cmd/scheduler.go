package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/queue"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/scheduler"
	"github.com/spf13/cobra"
)

var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Start scheduler",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := cmd.Flag("config").Value.String()

		config := requireConfig(configPath)
		logg, cleanupLogger := requireLogger(config.Logger)
		defer cleanupLogger()

		storage, cleanupStorage := requireEventStorage(config.Storage)
		defer cleanupStorage()

		q := queue.New(config.Queue.URI())
		if err := q.Connect(); err != nil {
			logg.Error("scheduler connect: " + err.Error())
			resultCode = 1
			return
		}

		producer, err := q.CreateProducer(config.Queue.Exchange)
		if err != nil {
			logg.Error("scheduler create producer: " + err.Error())
			resultCode = 1
			return
		}

		ctx, cancel := signal.NotifyContext(context.Background(),
			syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		defer cancel()

		s := scheduler.New(ctx)
		taskFactory := scheduler.NewTaskFactory(storage, producer)
		if err := defineTasks(config.Scheduler, taskFactory, s, logg); err != nil {
			logg.Error("scheduler define tasks: " + err.Error())
			resultCode = 1
			return
		}

		go func() {
			<-ctx.Done()
			s.Stop()

			if err := q.Close(); err != nil {
				logg.Error("scheduler: " + err.Error())
				resultCode = 1
			}
		}()

		logg.Info("starting scheduler...")
		s.Start()
	},
}

func wrapTaskWithLog(name string, t scheduler.Task, logg logger.Logger) scheduler.Task {
	return func(ctx context.Context) error {
		logg.Info(name + " task call")

		err := t(ctx)
		if err != nil {
			logg.Error("error with " + name + ": " + err.Error())
		}

		return err
	}
}

func defineTasks(cfg SchedulerConf, f *scheduler.TaskFactory, s *scheduler.Scheduler, logg logger.Logger) error {
	if err := s.AddTask(
		cfg.SendNotification,
		wrapTaskWithLog("send notification", f.CreateSendNotificationTask(time.Minute), logg),
	); err != nil {
		return fmt.Errorf("definition notify task: %w", err)
	}

	if err := s.AddTask(
		cfg.DeleteOld,
		wrapTaskWithLog("delete old", f.CreateDeleteOldEventsTask(time.Minute), logg),
	); err != nil {
		return fmt.Errorf("definition delete old task: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(schedulerCmd)
}
