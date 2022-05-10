package scheduler

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/queue"
	mockqueue "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/queue/mocks"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
	mockstorage "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	ctx               = context.Background()
	notDefaultContext = mock.MatchedBy(func(c context.Context) bool {
		return c != ctx
	})
	now = mock.MatchedBy(func(t time.Time) bool {
		return time.Until(time.Now()) < time.Minute
	})
	lastYear = mock.MatchedBy(func(t time.Time) bool {
		ly := time.Now().AddDate(-1, 0, 0)
		return t.Sub(ly) < time.Minute
	})
)

func event(id, userID int64, title string, time time.Time) *storage.Event {
	return &storage.Event{
		ID:        id,
		UserID:    userID,
		Title:     title,
		TimeStart: time,
	}
}

func TestSendNotificationTaskSuccess(t *testing.T) {
	t.Run("no events", func(t *testing.T) {
		p := &mockqueue.Producer{}
		s := &mockstorage.EventStorage{}

		s.On("FindUnNotified", notDefaultContext, now).Once().Return([]*storage.Event{}, nil)

		f := NewTaskFactory(s, p)
		task := f.CreateSendNotificationTask(time.Second)
		require.NoError(t, task(ctx))
	})

	t.Run("several events", func(t *testing.T) {
		p := &mockqueue.Producer{}
		s := &mockstorage.EventStorage{}

		events := make([]*storage.Event, 0, 3)
		events = append(events, event(1, 1, "test event name 1", time.Now()))
		events = append(events, event(2, 2, "test event name 2", time.Now()))
		events = append(events, event(3, 3, "test event name 2", time.Now()))
		s.On("FindUnNotified", notDefaultContext, now).Once().Return(events, nil)

		s.On("MarkNotified", notDefaultContext, []int64{1, 2, 3}).Once().Return(nil)

		p.On("Publish", mock.MatchedBy(func(m *queue.Message) bool {
			return m.Key == EventNotificationKey &&
				strings.Contains(string(m.Payload), "test event name 1") ||
				strings.Contains(string(m.Payload), "test event name 2") ||
				strings.Contains(string(m.Payload), "test event name 3")
		})).Times(3).Return(nil)

		f := NewTaskFactory(s, p)
		task := f.CreateSendNotificationTask(time.Second)
		require.NoError(t, task(ctx))
	})
}

func TestSendNotificationTaskError(t *testing.T) {
	t.Run("FindUnNotified", func(t *testing.T) {
		p := &mockqueue.Producer{}
		s := &mockstorage.EventStorage{}

		testErr := errors.New("test error")
		s.On("FindUnNotified", notDefaultContext, now).Once().Return(nil, testErr)

		f := NewTaskFactory(s, p)
		task := f.CreateSendNotificationTask(time.Second)
		err := task(ctx)
		require.ErrorIs(t, err, testErr)
	})

	t.Run("Publish", func(t *testing.T) {
		p := &mockqueue.Producer{}
		s := &mockstorage.EventStorage{}

		events := make([]*storage.Event, 0, 1)
		events = append(events, event(1, 1, "test event name 1", time.Now()))
		s.On("FindUnNotified", notDefaultContext, now).Once().Return(events, nil)

		testErr := errors.New("test error")
		p.On("Publish", mock.MatchedBy(func(m *queue.Message) bool {
			return m.Key == EventNotificationKey &&
				strings.Contains(string(m.Payload), "test event name 1")
		})).Once().Return(testErr)

		f := NewTaskFactory(s, p)
		task := f.CreateSendNotificationTask(time.Second)
		err := task(ctx)
		require.ErrorIs(t, err, testErr)
	})

	t.Run("MarkNotified", func(t *testing.T) {
		p := &mockqueue.Producer{}
		s := &mockstorage.EventStorage{}

		events := make([]*storage.Event, 0, 1)
		events = append(events, event(1, 1, "test event name 1", time.Now()))
		s.On("FindUnNotified", notDefaultContext, now).Once().Return(events, nil)

		testErr := errors.New("test error")
		s.On("MarkNotified", notDefaultContext, []int64{1}).Once().Return(testErr)

		p.On("Publish", mock.MatchedBy(func(m *queue.Message) bool {
			return m.Key == EventNotificationKey &&
				strings.Contains(string(m.Payload), "test event name 1")
		})).Once().Return(testErr)

		f := NewTaskFactory(s, p)
		task := f.CreateSendNotificationTask(time.Second)
		err := task(ctx)
		require.ErrorIs(t, err, testErr)
	})
}

func TestDeleteOldEventsTaskSuccess(t *testing.T) {
	p := &mockqueue.Producer{}
	s := &mockstorage.EventStorage{}

	s.On("DeleteOlderThan", notDefaultContext, lastYear).Once().Return(nil)

	f := NewTaskFactory(s, p)
	task := f.CreateDeleteOldEventsTask(time.Second)
	require.NoError(t, task(ctx))
}

func TestDeleteOldEventsTaskError(t *testing.T) {
	p := &mockqueue.Producer{}
	s := &mockstorage.EventStorage{}

	testErr := errors.New("test error")
	s.On("DeleteOlderThan", notDefaultContext, lastYear).Once().Return(testErr)

	f := NewTaskFactory(s, p)
	task := f.CreateDeleteOldEventsTask(time.Second)
	err := task(ctx)
	require.ErrorIs(t, err, testErr)
}
