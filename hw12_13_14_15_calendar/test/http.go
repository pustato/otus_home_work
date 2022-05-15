package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	httpserver "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/server/http"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var ctx = context.Background()

type HTTPTestSuite struct {
	suite.Suite
	events []int64
	client http.Client
	host   string
}

func (s *HTTPTestSuite) addEventID(id int64) {
	s.events = append(s.events, id)
}

func (s *HTTPTestSuite) createEvent(r *httpserver.CreateEventRequest) int64 {
	s.T().Helper()

	body, err := json.Marshal(r)
	require.NoError(s.T(), err)

	rq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.host+"/event", strings.NewReader(string(body)))
	require.NoError(s.T(), err)

	rsp, err := s.client.Do(rq)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, rsp.StatusCode)

	rs := &httpserver.CreateEventResponse{}
	errString := s.parseHTTPResponse(rsp, rs)
	require.NoError(s.T(), rsp.Body.Close())
	require.Equal(s.T(), "null", errString)

	s.addEventID(rs.ID)

	return rs.ID
}

func (s *HTTPTestSuite) deleteEvent(id int64) {
	s.T().Helper()

	url := fmt.Sprintf("%s/event/%d", s.host, id)

	rq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	require.NoError(s.T(), err)

	rsp, err := s.client.Do(rq)
	require.NoError(s.T(), err)
	require.NoError(s.T(), rsp.Body.Close())
	require.Equal(s.T(), http.StatusNoContent, rsp.StatusCode)
}

func (s *HTTPTestSuite) getEvent(id int64) *httpserver.EventResponse {
	s.T().Helper()

	url := fmt.Sprintf("%s/event/%d", s.host, id)

	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(s.T(), err)

	rsp, err := s.client.Do(rq)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, rsp.StatusCode)

	event := &httpserver.EventResponse{}
	errString := s.parseHTTPResponse(rsp, event)
	require.NoError(s.T(), rsp.Body.Close())
	require.Equal(s.T(), "null", errString)

	return event
}

func (s *HTTPTestSuite) TearDownSuite() {
	s.T().Run("cleanup database", func(t *testing.T) {
		for _, id := range s.events {
			s.deleteEvent(id)
		}
	})
}

func (s *HTTPTestSuite) TestCRUD() {
	baseTime := getBaseTime().AddDate(1, 0, 0)
	var eventID int64
	s.T().Run("create event", func(t *testing.T) {
		eventID = s.createEvent(&httpserver.CreateEventRequest{
			UserID:      10,
			Title:       "Test CRUD",
			Description: "",
			TimeStart:   baseTime.Format(time.RFC3339),
			TimeEnd:     baseTime.AddDate(0, 0, 1).Format(time.RFC3339),
			Notify:      "0",
		})
	})

	s.T().Run("get event", func(t *testing.T) {
		event := s.getEvent(eventID)

		require.Equal(t, eventID, event.ID)
		require.Equal(t, int64(10), event.UserID)
		require.Equal(t, "Test CRUD", event.Title)
		require.Equal(t, baseTime.Format(time.RFC3339), event.TimeStart)
		require.Equal(t, baseTime.AddDate(0, 0, 1).Format(time.RFC3339), event.TimeEnd)
		require.Nil(t, event.NotifyAt)
	})

	s.T().Run("update", func(t *testing.T) {
		url := fmt.Sprintf("%s/event/%d", s.host, eventID)

		body, err := json.Marshal(httpserver.UpdateEventRequest{
			Title:       "Test CRUD updated",
			Description: "New Description",
			TimeStart:   baseTime.AddDate(0, 1, 0).Format(time.RFC3339),
			TimeEnd:     baseTime.AddDate(0, 1, 1).Format(time.RFC3339),
			Notify:      "1h",
		})
		require.NoError(t, err)

		rq, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(body)))
		require.NoError(t, err)

		rsp, err := s.client.Do(rq)
		require.NoError(t, err)
		require.NoError(t, rsp.Body.Close())
		require.Equal(t, http.StatusNoContent, rsp.StatusCode)

		event := s.getEvent(eventID)

		require.Equal(t, eventID, event.ID)
		require.Equal(t, int64(10), event.UserID)
		require.Equal(t, "Test CRUD updated", event.Title)
		require.Equal(t, "New Description", event.Description)
		require.Equal(t, baseTime.AddDate(0, 1, 0).Format(time.RFC3339), event.TimeStart)
		require.Equal(t, baseTime.AddDate(0, 1, 1).Format(time.RFC3339), event.TimeEnd)
		require.Equal(t, baseTime.AddDate(0, 1, 0).Add(-time.Hour).Format(time.RFC3339), *event.NotifyAt)
	})

	s.T().Run("delete event", func(t *testing.T) {
		s.deleteEvent(eventID)
	})
}

func (s *HTTPTestSuite) TestSearch() {
	s.T().Run("create events", func(t *testing.T) {
		for _, te := range getTestEvents() {
			s.createEvent(&httpserver.CreateEventRequest{
				UserID:      int64(te.UserID),
				Title:       te.Title,
				Description: te.Description,
				TimeStart:   te.TimeStart.Format(time.RFC3339),
				TimeEnd:     te.TimeEnd.Format(time.RFC3339),
				Notify:      te.Notify,
			})
		}
	})

	s.T().Run("query for period", func(t *testing.T) {
		date := getBaseTime()
		testData := []struct {
			period         string
			userID         string
			expectedCount  int
			date           time.Time
			descriptionHas string
		}{
			{"day", "2", 2, date, "@find-for-day"},
			{"week", "1", 4, date, "@find-for-week"},
			{"month", "1", 6, date, "@find-for-month"},
			{"month", "1", 2, date.AddDate(0, 1, 1), "@find-for-next-month"},
		}

		for i, td := range testData {
			td := td
			t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
				rq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.host+"/events/"+td.period, nil)
				require.NoError(t, err)

				q := rq.URL.Query()
				q.Add("date", td.date.Format(time.RFC3339))
				q.Add("userId", td.userID)

				rq.URL.RawQuery = q.Encode()

				rsp, err := s.client.Do(rq)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, rsp.StatusCode)

				var events []httpserver.EventResponse
				errString := s.parseHTTPResponse(rsp, &events)
				require.NoError(t, rsp.Body.Close())
				require.Equal(t, "null", errString)

				require.Equal(t, td.expectedCount, len(events))
				for _, e := range events {
					require.Contains(t, e.Description, td.descriptionHas)
				}
			})
		}
	})
}

func (s *HTTPTestSuite) TestDeleteOld() {
	s.T().Run("create too old event and ensure it will be deleted", func(t *testing.T) {
		baseTime := time.Now().AddDate(-1, 0, -1)

		eventID := s.createEvent(&httpserver.CreateEventRequest{
			UserID:      10,
			Title:       "Too old event",
			Description: "",
			TimeStart:   baseTime.Format(time.RFC3339),
			TimeEnd:     baseTime.Add(time.Hour).Format(time.RFC3339),
			Notify:      "0",
		})

		_ = s.getEvent(eventID)

		url := fmt.Sprintf("%s/event/%d", s.host, eventID)
		require.Eventually(t, func() bool {
			rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			require.NoError(t, err)

			rsp, err := s.client.Do(rq)
			require.NoError(t, err)
			require.NoError(t, rsp.Body.Close())

			return rsp.StatusCode == http.StatusNotFound
		}, time.Second*30, 250*time.Millisecond)
	})
}

func (s *HTTPTestSuite) TestBusinessErrors() {
	baseTime := getBaseTime()

	s.createEvent(&httpserver.CreateEventRequest{
		UserID:      10,
		Title:       "Event for 'time is busy' error",
		Description: "",
		TimeStart:   baseTime.AddDate(0, 1, 0).Format(time.RFC3339),
		TimeEnd:     baseTime.AddDate(0, 1, 0).Add(time.Hour).Format(time.RFC3339),
		Notify:      "0",
	})

	testData := []struct {
		rq        httpserver.CreateEventRequest
		errString string
	}{
		{
			rq: httpserver.CreateEventRequest{
				UserID:      10,
				Title:       strings.Repeat("a", 101),
				Description: "",
				TimeStart:   baseTime.Format(time.RFC3339),
				TimeEnd:     baseTime.Add(time.Hour).Format(time.RFC3339),
				Notify:      "0",
			},
			errString: "title is too long",
		},
		{
			rq: httpserver.CreateEventRequest{
				UserID:      10,
				Title:       "Test business errors event",
				Description: "",
				TimeStart:   baseTime.Add(time.Hour).Format(time.RFC3339),
				TimeEnd:     baseTime.Format(time.RFC3339),
				Notify:      "0",
			},
			errString: "time end must be greater than time start",
		},
		{
			rq: httpserver.CreateEventRequest{
				UserID:      10,
				Title:       "Test business errors event",
				Description: "",
				TimeStart:   baseTime.Format(time.RFC3339),
				TimeEnd:     baseTime.Format(time.RFC3339),
				Notify:      "0",
			},
			errString: "time end must be greater than time start",
		},
		{
			rq: httpserver.CreateEventRequest{
				UserID:      10,
				Title:       "Test business errors event",
				Description: "",
				TimeStart:   baseTime.AddDate(0, 1, 0).Format(time.RFC3339),
				TimeEnd:     baseTime.AddDate(0, 1, 0).Add(time.Hour).Format(time.RFC3339),
				Notify:      "0",
			},
			errString: "time is busy",
		},
	}

	for i, td := range testData {
		td := td
		s.T().Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			body, err := json.Marshal(td.rq)
			require.NoError(t, err)

			rq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.host+"/event", strings.NewReader(string(body)))
			require.NoError(t, err)

			rsp, err := s.client.Do(rq)
			require.NoError(t, err)
			require.Equal(t, http.StatusUnprocessableEntity, rsp.StatusCode)

			body, err = io.ReadAll(rsp.Body)
			require.NoError(t, err)
			require.NoError(t, rsp.Body.Close())
			require.Contains(t, string(body), td.errString)
		})
	}
}

func (s *HTTPTestSuite) parseHTTPResponse(rsp *http.Response, v interface{}) string {
	body, err := io.ReadAll(rsp.Body)
	require.NoError(s.T(), err)

	var obj map[string]json.RawMessage
	err = json.Unmarshal(body, &obj)
	require.NoError(s.T(), err)

	data, ok := obj["data"]
	require.True(s.T(), ok)
	err = json.Unmarshal(data, v)
	require.NoError(s.T(), err)

	errString, ok := obj["error"]
	require.True(s.T(), ok)
	return string(errString)
}
