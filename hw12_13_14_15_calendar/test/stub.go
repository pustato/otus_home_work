package test

import (
	"time"

	"github.com/jinzhu/now"
)

type eventStub struct {
	UserID      int
	Title       string
	Description string
	TimeStart   time.Time
	TimeEnd     time.Time
	Notify      string
}

func getBaseTime() time.Time {
	return now.BeginningOfMonth().UTC()
}

func getTestEvents() []*eventStub {
	base := getBaseTime()

	return []*eventStub{
		// week events
		{
			UserID:      1,
			Title:       "Event #1",
			Description: "Description #1 @find-for-week @find-for-month",
			TimeStart:   base,
			TimeEnd:     base.Add(time.Hour),
			Notify:      "10m",
		},
		{
			UserID:      1,
			Title:       "Event #2",
			Description: "Description #2 @find-for-week @find-for-month",
			TimeStart:   base.AddDate(0, 0, 1),
			TimeEnd:     base.AddDate(0, 0, 1).Add(time.Hour),
			Notify:      "10m",
		},
		{
			UserID:      1,
			Title:       "Event #3",
			Description: "Description #3 @find-for-week @find-for-month",
			TimeStart:   base.AddDate(0, 0, 2),
			TimeEnd:     base.AddDate(0, 0, 2).Add(time.Hour),
			Notify:      "10m",
		},
		{
			UserID:      1,
			Title:       "Event #4",
			Description: "Description #4 @find-for-week @find-for-month",
			TimeStart:   base.AddDate(0, 0, 3),
			TimeEnd:     base.AddDate(0, 0, 3).Add(time.Hour),
			Notify:      "10m",
		},

		// month events
		{
			UserID:      1,
			Title:       "Event #5",
			Description: "Description #5 @find-for-month",
			TimeStart:   base.AddDate(0, 0, 14),
			TimeEnd:     base.AddDate(0, 0, 14).Add(time.Hour),
			Notify:      "10m",
		},
		{
			UserID:      1,
			Title:       "Event #6",
			Description: "Description #6 @find-for-month",
			TimeStart:   base.AddDate(0, 0, 15),
			TimeEnd:     base.AddDate(0, 0, 15).Add(time.Hour),
			Notify:      "10m",
		},

		// next month events
		{
			UserID:      1,
			Title:       "Event #7",
			Description: "Description #7 @find-for-next-month",
			TimeStart:   base.AddDate(0, 1, 1),
			TimeEnd:     base.AddDate(0, 1, 1).Add(time.Hour),
			Notify:      "10m",
		},
		{
			UserID:      1,
			Title:       "Event #8",
			Description: "Description #8  @find-for-next-month",
			TimeStart:   base.AddDate(0, 1, 2),
			TimeEnd:     base.AddDate(0, 1, 2).Add(time.Hour),
			Notify:      "10m",
		},

		// another user events day event
		{
			UserID:      2,
			Title:       "Event #9",
			Description: "Description #9 @find-for-day",
			TimeStart:   base,
			TimeEnd:     base.Add(time.Hour),
			Notify:      "10m",
		},
		{
			UserID:      2,
			Title:       "Event #10",
			Description: "Description #10 @find-for-day",
			TimeStart:   base.Add(time.Hour * 2),
			TimeEnd:     base.Add(time.Hour * 3),
			Notify:      "10m",
		},
	}
}
