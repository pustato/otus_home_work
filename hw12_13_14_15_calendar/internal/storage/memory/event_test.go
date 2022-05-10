package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

var (
	testZeroTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx          = context.Background()
)

func gen(userID int64, title, description string, baseTime time.Time) *storage.Event {
	return &storage.Event{
		UserID:      userID,
		Title:       title,
		Description: description,
		TimeStart:   baseTime,
		TimeEnd:     baseTime.Add(time.Hour),
	}
}

func compare(t *testing.T, a, b *storage.Event) {
	t.Helper()

	require.Equal(t, a.ID, b.ID)
	require.Equal(t, a.UserID, b.UserID)
	require.Equal(t, a.Title, b.Title)
	require.Equal(t, a.Description, b.Description)
	require.Equal(t, a.TimeStart, b.TimeStart)
	require.Equal(t, a.TimeEnd, b.TimeEnd)
	require.Equal(t, a.CreatedAt, b.CreatedAt)
}

func TestStorage(t *testing.T) {
	t.Run("simple CRUD", func(t *testing.T) {
		unit := New()

		event := gen(1, "title", "descr", testZeroTime)
		event2 := gen(2, "title2", "descr2", testZeroTime.Add(24*time.Hour))

		id, err := unit.Create(ctx, event)
		require.NoError(t, err)
		require.Equal(t, int64(1), id)
		require.Equal(t, id, event.ID)

		id2, err := unit.Create(ctx, event2)
		require.NoError(t, err)
		require.Equal(t, int64(2), id2)
		require.Equal(t, id2, event2.ID)

		found, err := unit.GetByID(ctx, event.ID)
		require.NoError(t, err)
		require.Equal(t, event, found)

		found2, err := unit.GetByID(ctx, event2.ID)
		require.NoError(t, err)
		require.Equal(t, event2, found2)

		_, err = unit.GetByID(ctx, 3)
		require.ErrorIs(t, err, storage.ErrNotFound)

		event.UserID = 10
		event.Title = "new title"
		event.Description = "new description"
		event.TimeStart = testZeroTime.Add(time.Minute)
		event.TimeEnd = testZeroTime.Add(time.Minute)

		err = unit.Update(ctx, event)
		require.NoError(t, err)

		foundUpdated, err := unit.GetByID(ctx, event.ID)
		require.NoError(t, err)
		compare(t, event, foundUpdated)
		require.NotEqual(t, event.UpdatedAt, foundUpdated.UpdatedAt)

		require.NoError(t, unit.Delete(ctx, event.ID))
		require.NoError(t, unit.Delete(ctx, event2.ID))

		_, err = unit.GetByID(ctx, event.ID)
		require.ErrorIs(t, err, storage.ErrNotFound)
		_, err = unit.GetByID(ctx, event2.ID)
		require.ErrorIs(t, err, storage.ErrNotFound)
	})

	t.Run("filter by time", func(t *testing.T) {
		unit := New()
		for i := 1; i <= 4; i++ {
			timeStart := testZeroTime.Add(time.Hour * time.Duration(i))

			for u := int64(1); u <= 2; u++ {
				e := gen(u, "", "", timeStart)
				_, err := unit.Create(ctx, e)
				require.NoError(t, err)
			}
		}

		t.Run("simple case", func(t *testing.T) {
			events, err := unit.FindForInterval(ctx, 1, testZeroTime.Add(time.Minute), testZeroTime.Add(3*time.Hour+1), 99, 0)
			require.NoError(t, err)
			require.Len(t, events, 3)
		})

		t.Run("works like BETWEEN from SQL", func(t *testing.T) {
			events, err := unit.FindForInterval(ctx, 1, testZeroTime.Add(time.Hour), testZeroTime.Add(4*time.Hour), 99, 0)
			require.NoError(t, err)
			require.Len(t, events, 4)
		})

		t.Run("limit", func(t *testing.T) {
			events, err := unit.FindForInterval(ctx, 1, testZeroTime.Add(time.Hour), testZeroTime.Add(3*time.Hour), 2, 0)
			require.NoError(t, err)
			require.Len(t, events, 2)
		})

		t.Run("offset", func(t *testing.T) {
			events, err := unit.FindForInterval(ctx, 2, testZeroTime.Add(time.Hour), testZeroTime.Add(3*time.Hour), 99, 1)
			require.NoError(t, err)
			require.Len(t, events, 2)
		})
	})

	t.Run("events in storage and events in results are different instances", func(t *testing.T) {
		t.Run("storage contains copy on create", func(t *testing.T) {
			unit := New()
			original := gen(1, "title", "descr", testZeroTime)
			_, err := unit.Create(ctx, original)
			require.NoError(t, err)

			fromStorage, err := unit.GetByID(ctx, original.ID)
			require.NoError(t, err)
			compare(t, original, fromStorage)

			original.UserID++
			require.NotEqual(t, original, fromStorage)
		})

		t.Run("storage contains copy on update", func(t *testing.T) {
			unit := New()
			original := gen(1, "title", "descr", testZeroTime)
			_, err := unit.Create(ctx, original)
			require.NoError(t, err)

			original.UserID++
			err = unit.Update(ctx, original)
			require.NoError(t, err)

			fromStorage, err := unit.GetByID(ctx, original.ID)
			require.NoError(t, err)
			compare(t, original, fromStorage)

			original.UserID++
			require.NotEqual(t, original, fromStorage)
		})

		t.Run("storage contains copy of search", func(t *testing.T) {
			unit := New()
			original := gen(1, "", "", testZeroTime)
			_, err := unit.Create(ctx, original)
			require.NoError(t, err)

			chunk, err := unit.FindForInterval(ctx, 1, testZeroTime, testZeroTime, 1, 0)
			require.NoError(t, err)
			require.Len(t, chunk, 1)

			original.UserID++
			require.NotEqual(t, original, chunk[0])
		})
	})
}

func TestStorageConcurrency(t *testing.T) {
	t.Run("work concurrently", func(t *testing.T) {
		unit := New()
		eventsPerIteration := 50
		iterationsCount := 100

		doSomeWork := func(n int) {
			ids := make([]int64, 0, n)
			for i := 0; i < n; i++ {
				id, err := unit.Create(ctx, gen(1, "", "", testZeroTime))
				require.NoError(t, err)

				ids = append(ids, id)
			}

			for _, id := range ids {
				_, err := unit.GetByID(ctx, id)
				require.NoError(t, err)

				_, err = unit.FindForInterval(ctx, 1, testZeroTime, testZeroTime, 1, 0)
				require.NoError(t, err)

				require.NoError(t, unit.Delete(ctx, id))
			}
		}

		wg := sync.WaitGroup{}
		wg.Add(iterationsCount)
		for i := 0; i < iterationsCount; i++ {
			go func() {
				defer wg.Done()
				doSomeWork(eventsPerIteration)
			}()
		}
		wg.Wait()

		e := gen(1, "", "", testZeroTime)
		_, _ = unit.Create(ctx, e)
		require.Equal(t, int64(eventsPerIteration*iterationsCount+1), e.ID)
	})
}
