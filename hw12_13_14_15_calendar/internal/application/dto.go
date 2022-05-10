package application

import (
	"time"
)

type CreateDTO struct {
	UserID      int64
	Title       string
	Description string
	TimeStart   time.Time
	TimeEnd     time.Time
	Notify      time.Duration
}

type UpdateDTO struct {
	Title       string
	Description string
	TimeStart   time.Time
	TimeEnd     time.Time
	Notify      time.Duration
}

type FindByDateDTO struct {
	UserID int64
	Date   time.Time
	Limit  uint8
	Offset uint8
}
