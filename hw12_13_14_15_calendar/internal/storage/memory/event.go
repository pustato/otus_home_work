package memory

import (
	"context"
	"sync"
	"time"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
)

var _ storage.EventStorage = (*EventStorage)(nil)

type EventStorage struct {
	mu sync.RWMutex

	id     int64
	events map[int64]*storage.Event
}

func New() *EventStorage {
	return &EventStorage{
		events: make(map[int64]*storage.Event),
	}
}

func (s *EventStorage) Create(_ context.Context, event *storage.Event) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.id++
	event.ID = s.id
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	val := *event
	cpy := val
	s.events[s.id] = &cpy

	return s.id, nil
}

func (s *EventStorage) Update(ctx context.Context, event *storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, err := s.getByID(ctx, event.ID)
	if err != nil {
		// nolint:nilerr
		return nil
	}

	val := *event
	val.ID = e.ID
	val.UpdatedAt = time.Now()
	s.events[event.ID] = &val

	return nil
}

func (s *EventStorage) Delete(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.events, id)
	return nil
}

func (s *EventStorage) GetByID(ctx context.Context, id int64) (*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getByID(ctx, id)
}

func (s *EventStorage) getByID(_ context.Context, id int64) (*storage.Event, error) {
	event, ok := s.events[id]
	if !ok {
		return nil, storage.ErrNotFound
	}

	value := *event
	cpy := value
	return &cpy, nil
}

func (s *EventStorage) FindForInterval(
	_ context.Context,
	userID int64,
	from, to time.Time,
	limit, offset uint8) ([]*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*storage.Event, 0, limit)

	for _, e := range s.events {
		if !(e.UserID == userID &&
			(e.TimeStart.Equal(from) || e.TimeStart.After(from)) &&
			(e.TimeStart.Equal(to) || e.TimeStart.Before(to))) {
			continue
		}

		if offset != 0 {
			offset--
			continue
		}

		val := *e
		cpy := val
		result = append(result, &cpy)
		if len(result) == int(limit) {
			break
		}
	}

	return result, nil
}
