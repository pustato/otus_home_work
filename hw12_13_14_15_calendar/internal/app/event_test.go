package app

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jinzhu/now"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
	mockstorage "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const longTitle = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec nec enim aliquam, suscipit dui turpis.
`

var (
	ctx      = context.Background()
	anyEvent = mock.MatchedBy(func(e *storage.Event) bool {
		return true
	})
)

func eventStub(t *testing.T) storage.Event {
	t.Helper()

	timeStart := time.Now().Add(-24 * time.Hour)
	return storage.Event{
		ID:          1,
		UserID:      1,
		Title:       "t",
		Description: "d",
		TimeStart:   timeStart,
		TimeEnd:     timeStart.Add(time.Hour),
		NotifyAt:    storage.CreateNotificationTime(timeStart, time.Minute),
	}
}

func createDtoToEvent(t *testing.T, dto CreateDTO) storage.Event {
	t.Helper()

	return storage.Event{
		UserID:      dto.UserID,
		Title:       dto.Title,
		Description: dto.Description,
		TimeStart:   dto.TimeStart,
		TimeEnd:     dto.TimeEnd,
		NotifyAt:    storage.CreateNotificationTime(dto.TimeStart, dto.Notify),
	}
}

func TestEvents_Get(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		id := int64(32)

		expected := eventStub(t)
		storageMock := mockstorage.EventStorage{}
		storageMock.
			On("GetByID", ctx, id).
			Once().
			Return(&expected, nil)

		uc := Events{
			storage: &storageMock,
		}
		actual, err := uc.GetByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, &expected, actual)
	})

	t.Run("not found case", func(t *testing.T) {
		id := int64(91)

		storageMock := mockstorage.EventStorage{}
		storageMock.
			On("GetByID", ctx, id).
			Once().
			Return(nil, storage.ErrNotFound)

		uc := Events{
			storage: &storageMock,
		}
		actual, err := uc.GetByID(ctx, id)
		require.Nil(t, actual)
		require.ErrorIs(t, err, ErrEventIsNotExists)
	})

	t.Run("storage error case", func(t *testing.T) {
		id := int64(128)
		errTest := errors.New("some storage error")

		storageMock := mockstorage.EventStorage{}
		storageMock.
			On("GetByID", ctx, id).
			Once().
			Return(nil, errTest)

		uc := Events{
			storage: &storageMock,
		}
		actual, err := uc.GetByID(ctx, id)
		require.Nil(t, actual)
		require.ErrorIs(t, err, errTest)
	})
}

func TestEventUseCase_Create(t *testing.T) {
	noww := time.Now()

	t.Run("success case", func(t *testing.T) {
		testData := []CreateDTO{
			{1, "title", "", noww, noww.Add(time.Hour), 0},
			{1, "title", "descr", noww, noww.Add(time.Hour), 0},
			{1, "title", "descr", noww, noww.Add(time.Hour), time.Minute * 10},
		}

		for i, dto := range testData {
			dto := dto
			t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
				storageMock := mockstorage.EventStorage{}
				storageMock.
					On("FindForInterval", ctx, dto.UserID, dto.TimeStart, dto.TimeEnd, uint8(2), uint8(0)).
					Once().
					Return([]*storage.Event{}, nil)
				storageMock.
					On("Create", ctx, anyEvent).
					Once().
					Return(int64(32), nil)

				uc := Events{
					storage: &storageMock,
				}

				id, err := uc.Create(ctx, dto)
				require.NoError(t, err)
				require.Equal(t, int64(32), id)
			})
		}
	})

	t.Run("validation error", func(t *testing.T) {
		testData := []struct {
			dto CreateDTO
			err []error
		}{
			{
				dto: CreateDTO{1, longTitle, "", noww, noww.Add(time.Hour), 0},
				err: []error{ErrTitleTooLong},
			},
			{
				dto: CreateDTO{1, "title", "", noww, noww.Add(-time.Hour), 0},
				err: []error{ErrTimeEndMustBeGreaterThanStart},
			},
			{
				dto: CreateDTO{2, "title", "", noww, noww.Add(time.Hour), 0},
				err: []error{ErrTimeIsBusy},
			},
			{
				dto: CreateDTO{2, longTitle, "", noww, noww.Add(-time.Hour), 0},
				err: []error{ErrTitleTooLong, ErrTimeEndMustBeGreaterThanStart, ErrTimeIsBusy},
			},
		}

		for i, td := range testData {
			dto, expectedErrors := td.dto, td.err
			t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
				storageMock := mockstorage.EventStorage{}
				se := createDtoToEvent(t, dto)
				existed := se
				existed.ID = 99

				storageMock.
					On("FindForInterval", ctx, int64(1), dto.TimeStart, dto.TimeEnd, uint8(2), uint8(0)).
					Return([]*storage.Event{}, nil)
				storageMock.
					On("FindForInterval", ctx, int64(2), dto.TimeStart, dto.TimeEnd, uint8(2), uint8(0)).
					Return([]*storage.Event{&existed}, nil)

				uc := Events{
					storage: &storageMock,
				}
				var v *ValidationErrors
				id, err := uc.Create(ctx, dto)
				require.Equal(t, int64(0), id)
				require.ErrorAs(t, err, &v)
				require.Equal(t, len(v.Errors()), len(expectedErrors))

				matchCounter := 0
				for _, actual := range v.Errors() {
					for _, expected := range expectedErrors {
						if errors.Is(actual, expected) {
							matchCounter++
						}
					}
				}

				require.Equal(t, len(expectedErrors), matchCounter)
			})
		}
	})

	t.Run("storage error", func(t *testing.T) {
		dto := CreateDTO{1, "title", "", noww, noww.Add(time.Hour), 0}

		t.Run("find for interval", func(t *testing.T) {
			storageMock := mockstorage.EventStorage{}

			errTest := errors.New("some error")
			storageMock.
				On("FindForInterval", ctx, dto.UserID, dto.TimeStart, dto.TimeEnd, uint8(2), uint8(0)).
				Once().
				Return([]*storage.Event{}, errTest)

			uc := Events{
				storage: &storageMock,
			}

			id, err := uc.Create(ctx, dto)
			require.Equal(t, int64(0), id)
			require.ErrorIs(t, err, errTest)
		})

		t.Run("create", func(t *testing.T) {
			storageMock := mockstorage.EventStorage{}

			errTest := errors.New("some error")
			storageMock.
				On("FindForInterval", ctx, dto.UserID, dto.TimeStart, dto.TimeEnd, uint8(2), uint8(0)).
				Once().
				Return([]*storage.Event{}, nil)
			storageMock.
				On("Create", ctx, anyEvent).
				Once().
				Return(int64(0), errTest)

			uc := Events{
				storage: &storageMock,
			}

			id, err := uc.Create(ctx, dto)
			require.Equal(t, int64(0), id)
			require.ErrorIs(t, err, errTest)
		})
	})
}

func TestEventUseCase_Update(t *testing.T) {
	noww := time.Now()
	sampleEvent := eventStub(t)

	t.Run("success case", func(t *testing.T) {
		testData := []UpdateDTO{
			{"title", "", noww, noww.Add(time.Hour), 0},
			{"title", "description", noww, noww.Add(time.Hour), 0},
			{"title", "", noww, noww.Add(time.Hour), time.Minute},
			{"title", "description", noww, noww.Add(time.Hour), time.Minute},
		}

		userID := int64(1)
		for i, dto := range testData {
			dto := dto
			t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
				storageMock := mockstorage.EventStorage{}
				sampleEvent := sampleEvent

				storageMock.
					On("FindForInterval", ctx, userID, dto.TimeStart, dto.TimeEnd, uint8(2), uint8(0)).
					Once().
					Return([]*storage.Event{}, nil)
				storageMock.
					On("GetByID", ctx, int64(1)).
					Once().
					Return(&sampleEvent, nil)
				storageMock.
					On("Update", ctx, anyEvent).
					Once().
					Return(nil)

				uc := Events{
					storage: &storageMock,
				}

				err := uc.Update(ctx, 1, dto)
				require.NoError(t, err)
				require.Equal(t, sampleEvent.Title, dto.Title)
				require.Equal(t, sampleEvent.Description, dto.Description)
				require.Equal(t, sampleEvent.TimeStart, dto.TimeStart)
				require.Equal(t, sampleEvent.TimeEnd, dto.TimeEnd)
				require.Equal(t, sampleEvent.NotifyAt, storage.CreateNotificationTime(dto.TimeStart, dto.Notify))
			})
		}
	})

	t.Run("not found error", func(t *testing.T) {
		storageMock := mockstorage.EventStorage{}
		storageMock.
			On("GetByID", ctx, int64(1)).
			Once().
			Return(nil, storage.ErrNotFound)

		uc := Events{
			storage: &storageMock,
		}

		err := uc.Update(ctx, 1, UpdateDTO{})
		require.ErrorIs(t, err, ErrEventIsNotExists)
	})

	t.Run("storage error", func(t *testing.T) {
		errTest := errors.New("some storage error")

		t.Run("get by id", func(t *testing.T) {
			storageMock := mockstorage.EventStorage{}
			storageMock.
				On("GetByID", ctx, int64(1)).
				Once().
				Return(nil, errTest)

			uc := Events{
				storage: &storageMock,
			}

			err := uc.Update(ctx, 1, UpdateDTO{})
			require.ErrorIs(t, err, errTest)
		})

		t.Run("update", func(t *testing.T) {
			dto := UpdateDTO{"title", "", noww, noww.Add(time.Hour), 0}

			storageMock := mockstorage.EventStorage{}
			storageMock.
				On("GetByID", ctx, int64(1)).
				Once().
				Return(&sampleEvent, nil)
			storageMock.
				On("FindForInterval", ctx, sampleEvent.UserID, dto.TimeStart, dto.TimeEnd, uint8(2), uint8(0)).
				Once().
				Return([]*storage.Event{}, nil)
			storageMock.
				On("Update", ctx, anyEvent).
				Once().
				Return(errTest)

			uc := Events{
				storage: &storageMock,
			}

			err := uc.Update(ctx, 1, dto)
			require.ErrorIs(t, err, errTest)
		})
	})
}

func TestEventUseCase_Delete(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		id := int64(56)

		storageMock := mockstorage.EventStorage{}
		storageMock.
			On("Delete", ctx, id).
			Once().
			Return(nil)

		uc := Events{
			storage: &storageMock,
		}
		require.NoError(t, uc.Delete(ctx, id))
	})

	t.Run("error case", func(t *testing.T) {
		id := int64(56)
		err := errors.New("some storage error")

		storageMock := mockstorage.EventStorage{}
		storageMock.
			On("Delete", ctx, id).
			Once().
			Return(err)

		uc := Events{
			storage: &storageMock,
		}
		require.ErrorIs(t, uc.Delete(ctx, id), err)
	})
}

func TestEvents_FindForInterval(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		testData := []FindByDateDTO{
			{1, time.Now(), 10, 0},
			{2, time.Now().AddDate(0, 0, 32), 50, 10},
			{3, time.Now().AddDate(1, 1, 0), 100, 100},
		}

		s1, s2 := eventStub(t), eventStub(t)
		expected := make([]*storage.Event, 0, 2)
		expected = append(expected, &s1)
		expected = append(expected, &s2)

		for i, dto := range testData {
			dto := dto
			t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
				beginningOfDay := now.With(dto.Date).BeginningOfDay()
				endOfDay := now.With(dto.Date).EndOfDay()

				beginningOfWeek := now.With(dto.Date).BeginningOfWeek()
				endOfWeek := now.With(dto.Date).EndOfWeek()

				beginningOfMonth := now.With(dto.Date).BeginningOfMonth()
				endOfMonth := now.With(dto.Date).EndOfMonth()

				storageMock := mockstorage.EventStorage{}
				storageMock.
					On("FindForInterval", ctx, dto.UserID, beginningOfDay, endOfDay, dto.Limit, dto.Offset).
					Once().
					Return(expected, nil)
				storageMock.
					On("FindForInterval", ctx, dto.UserID, beginningOfWeek, endOfWeek, dto.Limit, dto.Offset).
					Once().
					Return(expected, nil)
				storageMock.
					On("FindForInterval", ctx, dto.UserID, beginningOfMonth, endOfMonth, dto.Limit, dto.Offset).
					Once().
					Return(expected, nil)

				uc := Events{
					storage: &storageMock,
				}
				actual, err := uc.FindForDay(ctx, dto)
				require.NoError(t, err)
				require.EqualValues(t, expected, actual)

				actual, err = uc.FindForWeek(ctx, dto)
				require.NoError(t, err)
				require.EqualValues(t, expected, actual)

				actual, err = uc.FindForMonth(ctx, dto)
				require.NoError(t, err)
				require.EqualValues(t, expected, actual)
			})
		}
	})

	t.Run("test storage error", func(t *testing.T) {
		errTest := errors.New("storage error")
		dto := FindByDateDTO{1, time.Now(), 10, 0}
		beginningOfDay := now.With(dto.Date).BeginningOfDay()
		endOfDay := now.With(dto.Date).EndOfDay()

		beginningOfWeek := now.With(dto.Date).BeginningOfWeek()
		endOfWeek := now.With(dto.Date).EndOfWeek()

		beginningOfMonth := now.With(dto.Date).BeginningOfMonth()
		endOfMonth := now.With(dto.Date).EndOfMonth()

		storageMock := mockstorage.EventStorage{}
		storageMock.
			On("FindForInterval", ctx, dto.UserID, beginningOfDay, endOfDay, dto.Limit, dto.Offset).
			Once().
			Return(nil, errTest)
		storageMock.
			On("FindForInterval", ctx, dto.UserID, beginningOfWeek, endOfWeek, dto.Limit, dto.Offset).
			Once().
			Return(nil, errTest)
		storageMock.
			On("FindForInterval", ctx, dto.UserID, beginningOfMonth, endOfMonth, dto.Limit, dto.Offset).
			Once().
			Return(nil, errTest)

		uc := Events{
			storage: &storageMock,
		}
		events, err := uc.FindForDay(ctx, dto)
		require.Nil(t, events)
		require.ErrorIs(t, err, errTest)

		events, err = uc.FindForWeek(ctx, dto)
		require.Nil(t, events)
		require.ErrorIs(t, err, errTest)

		events, err = uc.FindForMonth(ctx, dto)
		require.Nil(t, events)
		require.ErrorIs(t, err, errTest)
	})
}
