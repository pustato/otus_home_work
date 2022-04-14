package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
)

var _ storage.EventStorage = (*EventStorage)(nil)

type EventStorage struct {
	db *sqlx.DB
}

func New() *EventStorage {
	return &EventStorage{}
}

func (s *EventStorage) Create(ctx context.Context, event *storage.Event) (int64, error) {
	q := `
		INSERT INTO 
			events (user_id, title, description, time_start, time_end, notify_at, created_at, updated_at)
		VALUES 
			(:user_id, :title, :description, :time_start, :time_end, :notify_at, :created_at, :updated_at)
		RETURNING id
		;
`
	now := time.Now()

	res, err := s.db.NamedQueryContext(
		ctx,
		q,
		map[string]interface{}{
			"user_id":     event.UserID,
			"title":       event.Title,
			"description": event.Description,
			"time_start":  event.TimeStart,
			"time_end":    event.TimeEnd,
			"notify_at":   event.NotifyAt,
			"created_at":  now,
			"updated_at":  now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("event create: %w", err)
	}
	defer func() {
		_ = res.Close()
		_ = res.Err()
	}()

	res.Next()
	if err := res.Scan(&event.ID); err != nil {
		return 0, fmt.Errorf("event retrieve last insert id: %w", err)
	}

	event.CreatedAt, event.UpdatedAt = now, now

	return event.ID, nil
}

func (s *EventStorage) Update(ctx context.Context, event *storage.Event) error {
	q := `
		UPDATE 
			events 
		SET 
			user_id=:user_id,
			title=:title,
			description=:description,
			time_start=:time_start,
			time_end=:time_end,
			updated_at=:updated_at,
			notify_at=:notify_at
		WHERE
			id=:id
		;
`
	now := time.Now()

	_, err := s.db.NamedExecContext(
		ctx,
		q,
		map[string]interface{}{
			"user_id":     event.UserID,
			"title":       event.Title,
			"description": event.Description,
			"time_start":  event.TimeStart,
			"time_end":    event.TimeEnd,
			"notify_at":   event.NotifyAt,
			"updated_at":  now,
			"id":          event.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("event update: %w", err)
	}

	event.UpdatedAt = now

	return nil
}

func (s *EventStorage) Delete(ctx context.Context, id int64) error {
	q := `
		DELETE FROM 
			events 
		WHERE 
			id=:id
		;
`

	_, err := s.db.NamedExecContext(ctx, q, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return fmt.Errorf("event delete: %w", err)
	}

	return nil
}

func (s *EventStorage) GetByID(ctx context.Context, id int64) (*storage.Event, error) {
	q := `
		SELECT
			id, 
			user_id,
			title,
			description,
			time_start, 
			time_end,
			notify_at,
			created_at,
			updated_at
		FROM 
			events
		WHERE
			id=:id
		;
`
	e := storage.Event{}

	rows, err := s.db.NamedQueryContext(ctx, q, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return nil, fmt.Errorf("event get by id: %w", err)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	if !rows.Next() {
		return nil, storage.ErrNotFound
	}

	if err := rows.Scan(
		&e.ID,
		&e.UserID,
		&e.Title,
		&e.Description,
		&e.TimeStart,
		&e.TimeEnd,
		&e.NotifyAt,
		&e.CreatedAt,
		&e.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("event get by id: %w", err)
	}

	return &e, nil
}

func (s *EventStorage) FindForInterval(
	ctx context.Context,
	userID int64,
	from, to time.Time,
	limit, offset uint8) ([]*storage.Event, error) {
	q := `
		SELECT
			id, 
			user_id,
			title,
			description,
			time_start, 
			time_end,
			notify_at,
			created_at,
			updated_at
		FROM
			events
		WHERE
			user_id=:user_id
			AND time_start BETWEEN :from AND :to
		ORDER BY time_start
		LIMIT :limit OFFSET :offset
		;
`
	rows, err := s.db.NamedQueryContext(ctx, q, map[string]interface{}{
		"user_id": userID,
		"from":    from,
		"to":      to,
		"limit":   limit,
		"offset":  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("event find for interval: %w", err)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	result := make([]*storage.Event, 0)

	for rows.Next() {
		e := storage.Event{}
		if err := rows.Scan(
			&e.ID,
			&e.UserID,
			&e.Title,
			&e.Description,
			&e.TimeStart,
			&e.TimeEnd,
			&e.NotifyAt,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("storage find for interval: scan: %w", err)
		}

		result = append(result, &e)
	}

	return result, nil
}

func (s *EventStorage) Connect(ctx context.Context, dsn string) error {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db connection with pgx: %w", err)
	}

	s.db = db
	return s.db.PingContext(ctx)
}

func (s *EventStorage) Close() error {
	return s.db.Close()
}
