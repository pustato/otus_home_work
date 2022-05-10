package application

import (
	"context"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
)

type EventsUseCase interface {
	Create(ctx context.Context, dto CreateDTO) (int64, error)
	Update(ctx context.Context, id int64, dto UpdateDTO) error
	Delete(ctx context.Context, id int64) error
	FindForDay(ctx context.Context, dto FindByDateDTO) ([]*storage.Event, error)
	FindForWeek(ctx context.Context, dto FindByDateDTO) ([]*storage.Event, error)
	FindForMonth(ctx context.Context, dto FindByDateDTO) ([]*storage.Event, error)
}

func NewEventUseCase(storage storage.EventStorage) EventsUseCase {
	return &Events{
		storage: storage,
	}
}
