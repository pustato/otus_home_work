package scheduler

import (
	"context"
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/queue"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
)

const EventNotificationKey = "event_notification"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type EventNotification struct {
	EventID   int64     `json:"eventId"`
	UserID    int64     `json:"userId"`
	Title     string    `json:"title"`
	TimeStart time.Time `json:"timeStart"`
}

type Task func(ctx context.Context) error

type TaskFactory struct {
	storage  storage.EventStorage
	producer queue.Producer
}

func (f *TaskFactory) CreateSendNotificationTask(timeout time.Duration) Task {
	return func(parent context.Context) error {
		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()

		events, err := f.storage.FindUnNotified(ctx, time.Now())
		if err != nil {
			return fmt.Errorf("notification task: %w", err)
		}

		if len(events) == 0 {
			return nil
		}

		ids := make([]int64, 0, len(events))
		for _, e := range events {
			n := &EventNotification{
				EventID:   e.ID,
				UserID:    e.UserID,
				Title:     e.Title,
				TimeStart: e.TimeStart,
			}

			payload, err := json.Marshal(n)
			if err != nil {
				return fmt.Errorf("notification task: %w", err)
			}

			if err := f.producer.Publish(&queue.Message{
				Key:     EventNotificationKey,
				Payload: payload,
			}); err != nil {
				return fmt.Errorf("notification task: %w", err)
			}

			ids = append(ids, e.ID)
		}

		if err := f.storage.MarkNotified(ctx, ids); err != nil {
			return fmt.Errorf("notification task: %w", err)
		}

		return nil
	}
}

func (f *TaskFactory) CreateDeleteOldEventsTask(timeout time.Duration) Task {
	return func(parent context.Context) error {
		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()

		lastYear := time.Now().AddDate(-1, 0, 0)
		err := f.storage.DeleteOlderThan(ctx, lastYear)
		if err != nil {
			return fmt.Errorf("delete old events task: %w", err)
		}

		return nil
	}
}

func NewTaskFactory(s storage.EventStorage, p queue.Producer) *TaskFactory {
	return &TaskFactory{
		s, p,
	}
}
