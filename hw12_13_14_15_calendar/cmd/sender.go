package cmd

import (
	"context"
	"os/signal"
	"syscall"

	jsoniter "github.com/json-iterator/go"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/queue"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/scheduler"
	"github.com/spf13/cobra"
)

const eventsQueueName = "events"

var senderCmd = &cobra.Command{
	Use:   "sender",
	Short: "start sender",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := cmd.Flag("config").Value.String()

		config := requireConfig(configPath)
		logg, cleanupLogger := requireLogger(config.Logger)
		defer cleanupLogger()

		q := queue.New(config.Queue.URI())
		if err := q.Connect(); err != nil {
			logg.Error("sender connect: " + err.Error())
			resultCode = 1
			return
		}

		consumer, err := q.CreateConsumer(config.Queue.Exchange, eventsQueueName, scheduler.EventNotificationKey)
		if err != nil {
			logg.Error("sender create consumer: " + err.Error())
			resultCode = 1
			return
		}

		ctx, cancel := signal.NotifyContext(context.Background(),
			syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		defer cancel()

		go func() {
			<-ctx.Done()

			logg.Info("stopping sender")
			if err := q.Close(); err != nil {
				logg.Error("sender close queue: " + err.Error())
				resultCode = 1
			}
		}()

		logg.Info("sender started...")
		if err := consumer.Consume(func(m *queue.Message) {
			logg.Info("incoming message with key=" + m.Key)
			switch m.Key {
			case scheduler.EventNotificationKey:
				handleEventNotification(logg, m)
			default:
				logg.Warn("unknown message")
			}

			logg.Info(string(m.Payload))
		}); err != nil {
			logg.Error("sender consume: " + err.Error())
			resultCode = 1
			return
		}
	},
}

func handleEventNotification(logg logger.Logger, m *queue.Message) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	n := &scheduler.EventNotification{}
	if err := json.Unmarshal(m.Payload, n); err != nil {
		logg.Error("sender event notification unmarshal: " + err.Error())
	}

	logg.Info("event notification received",
		"EventId", n.EventID,
		"UserID", n.UserID,
		"Title", n.Title,
		"TimeStart", n.TimeStart,
	)
}

func init() {
	rootCmd.AddCommand(senderCmd)
}
