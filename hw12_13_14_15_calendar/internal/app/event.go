package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/jinzhu/now"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
)

const MaxEventTitleLength = 100

var _ EventsUseCase = (*Events)(nil)

type Events struct {
	storage storage.EventStorage
}

func (c *Events) GetByID(ctx context.Context, id int64) (*storage.Event, error) {
	e, err := c.storage.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrEventIsNotExists
		}

		return nil, fmt.Errorf("event use case get: %w", err)
	}

	return e, nil
}

func (c *Events) Create(ctx context.Context, dto CreateDTO) (int64, error) {
	e := &storage.Event{
		UserID:      dto.UserID,
		Title:       dto.Title,
		Description: dto.Description,
		TimeStart:   dto.TimeStart,
		TimeEnd:     dto.TimeEnd,
		NotifyAt:    storage.CreateNotificationTime(dto.TimeStart, dto.Notify),
	}

	if err := c.validate(ctx, e); err != nil {
		return 0, err
	}

	id, err := c.storage.Create(ctx, e)
	if err != nil {
		return 0, fmt.Errorf("event use case create: %w", err)
	}

	return id, nil
}

func (c *Events) Update(ctx context.Context, id int64, dto UpdateDTO) error {
	e, err := c.storage.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrEventIsNotExists
		}

		return fmt.Errorf("event use case update: %w", err)
	}

	e.Title = dto.Title
	e.Description = dto.Description
	e.TimeStart = dto.TimeStart
	e.TimeEnd = dto.TimeEnd
	e.NotifyAt = storage.CreateNotificationTime(dto.TimeStart, dto.Notify)

	if err := c.validate(ctx, e); err != nil {
		return err
	}

	if err := c.storage.Update(ctx, e); err != nil {
		return fmt.Errorf("event use case update: %w", err)
	}

	return nil
}

func (c *Events) Delete(ctx context.Context, id int64) error {
	if err := c.storage.Delete(ctx, id); err != nil {
		return fmt.Errorf("event use case delete: %w", err)
	}

	return nil
}

func (c *Events) FindForDay(ctx context.Context, dto FindByDateDTO) ([]*storage.Event, error) {
	noww := now.With(dto.Date)

	from := noww.BeginningOfDay()
	to := noww.EndOfDay()

	events, err := c.storage.FindForInterval(ctx, dto.UserID, from, to, dto.Limit, dto.Offset)
	if err != nil {
		return nil, fmt.Errorf("event use case find for day: %w", err)
	}

	return events, nil
}

func (c *Events) FindForWeek(ctx context.Context, dto FindByDateDTO) ([]*storage.Event, error) {
	noww := now.With(dto.Date)

	from := noww.BeginningOfWeek()
	to := noww.EndOfWeek()

	events, err := c.storage.FindForInterval(ctx, dto.UserID, from, to, dto.Limit, dto.Offset)
	if err != nil {
		return nil, fmt.Errorf("event use case find for week: %w", err)
	}

	return events, nil
}

func (c *Events) FindForMonth(ctx context.Context, dto FindByDateDTO) ([]*storage.Event, error) {
	noww := now.With(dto.Date)

	from := noww.BeginningOfMonth()
	to := noww.EndOfMonth()

	events, err := c.storage.FindForInterval(ctx, dto.UserID, from, to, dto.Limit, dto.Offset)
	if err != nil {
		return nil, fmt.Errorf("event use case find for month: %w", err)
	}

	return events, nil
}

func (c *Events) validate(ctx context.Context, e *storage.Event) error {
	errs := make([]error, 0)

	if len(e.Title) > MaxEventTitleLength {
		errs = append(errs, fmt.Errorf("title lengts is %d/%d: %w", len(e.Title), MaxEventTitleLength, ErrTitleTooLong))
	}

	if e.TimeStart.After(e.TimeEnd) {
		errs = append(errs, ErrTimeEndMustBeGreaterThanStart)
	}

	existed, err := c.storage.FindForInterval(ctx, e.UserID, e.TimeStart, e.TimeEnd, 2, 0)
	if err != nil {
		return fmt.Errorf("validate event repository error: %w", err)
	}

	if len(existed) > 0 {
		for _, ex := range existed {
			if e.ID != ex.ID {
				errs = append(errs, ErrTimeIsBusy)
				break
			}
		}
	}

	if len(errs) > 0 {
		return &ValidationErrors{errors: errs}
	}

	return nil
}
