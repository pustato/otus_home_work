package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type EventStorage interface {
	Create(ctx context.Context, event *Event) (int64, error)
	Update(ctx context.Context, event *Event) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Event, error)
	FindForInterval(ctx context.Context,
		userID int64,
		from, to time.Time,
		limit, offset uint8) ([]*Event, error)
	FindUnNotified(ctx context.Context, t time.Time) ([]*Event, error)
	MarkNotified(ctx context.Context, ids []int64) error
	DeleteOlderThan(ctx context.Context, t time.Time) error
}

type NotificationTime = sql.NullTime

var ErrNotFound = errors.New("not found")

type Event struct {
	ID               int64
	UserID           int64
	Title            string
	Description      string
	TimeStart        time.Time
	TimeEnd          time.Time
	NotifyAt         NotificationTime
	CreatedAt        time.Time
	UpdatedAt        time.Time
	NotificationSent bool
}

func CreateNotificationTime(base time.Time, d time.Duration) NotificationTime {
	var t NotificationTime

	if d > 0 {
		t.Valid = true
		t.Time = base.Add(-d)
	}

	return t
}
