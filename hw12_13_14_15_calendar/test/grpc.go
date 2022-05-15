package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCTestSuite struct {
	suite.Suite
	events []int64
	client pb.CalendarClient
}

func (s *GRPCTestSuite) addEventID(id int64) {
	s.events = append(s.events, id)
}

func (s *GRPCTestSuite) TearDownSuite() {
	s.T().Run("cleanup database", func(t *testing.T) {
		for _, id := range s.events {
			_, err := s.client.DeleteEvent(ctx, &pb.EventRequest{
				Id: id,
			})
			require.NoError(t, err)
		}
	})
}

func (s *GRPCTestSuite) TestCRUD() {
	var eventID int64
	baseTime := getBaseTime().AddDate(1, 0, 0)

	s.T().Run("create event", func(t *testing.T) {
		rsp, err := s.client.CreateEvent(ctx, &pb.CreateEventRequest{
			UserId:      11,
			Title:       "Test CRUD",
			Description: "",
			TimeStart:   timestamppb.New(baseTime),
			TimeEnd:     timestamppb.New(baseTime.Add(time.Hour)),
			Notify:      nil,
		})
		require.NoError(t, err)
		s.addEventID(rsp.Id)

		eventID = rsp.Id
	})

	s.T().Run("get event", func(t *testing.T) {
		rsp, err := s.client.GetEvent(ctx, &pb.EventRequest{
			Id: eventID,
		})
		require.NoError(t, err)

		require.Equal(t, int64(11), rsp.UserId)
		require.Equal(t, "Test CRUD", rsp.Title)
		require.Equal(t, baseTime, rsp.TimeStart.AsTime())
		require.Equal(t, baseTime.Add(time.Hour), rsp.TimeEnd.AsTime())
		require.False(t, rsp.NotifyAt.Valid)
	})

	s.T().Run("update event", func(t *testing.T) {
		_, err := s.client.UpdateEvent(ctx, &pb.UpdateEventRequest{
			Id:          eventID,
			Title:       "Test CRUD updated",
			Description: "Updated description",
			TimeStart:   timestamppb.New(baseTime.AddDate(0, 1, 0)),
			TimeEnd:     timestamppb.New(baseTime.AddDate(0, 1, 1)),
			Notify:      durationpb.New(time.Minute * 10),
		})
		require.NoError(t, err)

		rsp, err := s.client.GetEvent(ctx, &pb.EventRequest{
			Id: eventID,
		})
		require.NoError(t, err)

		require.Equal(t, "Test CRUD updated", rsp.Title)
		require.Equal(t, "Updated description", rsp.Description)
		require.Equal(t, baseTime.AddDate(0, 1, 0), rsp.TimeStart.AsTime())
		require.Equal(t, baseTime.AddDate(0, 1, 1), rsp.TimeEnd.AsTime())
		require.True(t, rsp.NotifyAt.Valid)
	})

	s.T().Run("delete event", func(t *testing.T) {
		_, err := s.client.DeleteEvent(ctx, &pb.EventRequest{
			Id: eventID,
		})
		require.NoError(t, err)
	})
}

func (s *GRPCTestSuite) TestSearch() {
	s.T().Run("create events", func(t *testing.T) {
		for _, td := range getTestEvents() {
			rsp, err := s.client.CreateEvent(ctx, &pb.CreateEventRequest{
				UserId:      int64(td.UserID),
				Title:       td.Title,
				Description: td.Description,
				TimeStart:   timestamppb.New(td.TimeStart),
				TimeEnd:     timestamppb.New(td.TimeEnd),
				Notify:      nil,
			})
			require.NoError(t, err)
			s.addEventID(rsp.Id)
		}
	})

	s.T().Run("query for period", func(t *testing.T) {
		date := getBaseTime()
		testData := []struct {
			userID         int64
			expectedCount  int
			date           time.Time
			descriptionHas string
			f              func(r *pb.PeriodRequest) (*pb.EventCollection, error)
		}{
			{
				2, 2, date, "@find-for-day",
				func(r *pb.PeriodRequest) (*pb.EventCollection, error) {
					return s.client.FindForDay(ctx, r)
				},
			},
			{
				1, 4, date, "@find-for-week",
				func(r *pb.PeriodRequest) (*pb.EventCollection, error) {
					return s.client.FindForWeek(ctx, r)
				},
			},
			{
				1, 6, date, "@find-for-month",
				func(r *pb.PeriodRequest) (*pb.EventCollection, error) {
					return s.client.FindForMonth(ctx, r)
				},
			},
			{
				1, 2, date.AddDate(0, 1, 1), "@find-for-next-month",
				func(r *pb.PeriodRequest) (*pb.EventCollection, error) {
					return s.client.FindForMonth(ctx, r)
				},
			},
		}

		for i, td := range testData {
			td := td
			s.T().Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
				rq := &pb.PeriodRequest{
					UserId: td.userID,
					Date:   timestamppb.New(td.date),
					Limit:  100,
					Offset: 0,
				}

				events, err := td.f(rq)
				require.NoError(t, err)
				require.Equal(t, td.expectedCount, len(events.Events))

				for _, e := range events.Events {
					require.Contains(t, e.Description, td.descriptionHas)
				}
			})
		}
	})
}

func (s *GRPCTestSuite) TestBusinessErrors() {
	baseTime := getBaseTime()

	rsp, err := s.client.CreateEvent(ctx, &pb.CreateEventRequest{
		UserId:      11,
		Title:       "Event for 'time is busy' error",
		Description: "",
		TimeStart:   timestamppb.New(baseTime.AddDate(0, 1, 0)),
		TimeEnd:     timestamppb.New(baseTime.AddDate(0, 1, 0).Add(time.Hour)),
		Notify:      nil,
	})
	require.NoError(s.T(), err)
	s.addEventID(rsp.Id)

	testData := []struct {
		rq        *pb.CreateEventRequest
		errString string
	}{
		{
			rq: &pb.CreateEventRequest{
				UserId:      11,
				Title:       strings.Repeat("a", 101),
				Description: "",
				TimeStart:   timestamppb.New(baseTime),
				TimeEnd:     timestamppb.New(baseTime.Add(time.Hour)),
				Notify:      nil,
			},
			errString: "title is too long",
		},
		{
			rq: &pb.CreateEventRequest{
				UserId:      11,
				Title:       "Test business errors event",
				Description: "",
				TimeStart:   timestamppb.New(baseTime),
				TimeEnd:     timestamppb.New(baseTime),
				Notify:      nil,
			},
			errString: "time end must be greater than time start",
		},
		{
			rq: &pb.CreateEventRequest{
				UserId:      11,
				Title:       "Test business errors event",
				Description: "",
				TimeStart:   timestamppb.New(baseTime.Add(time.Hour)),
				TimeEnd:     timestamppb.New(baseTime),
				Notify:      nil,
			},
			errString: "time end must be greater than time start",
		},
		{
			rq: &pb.CreateEventRequest{
				UserId:      11,
				Title:       "Test business errors event",
				Description: "",
				TimeStart:   timestamppb.New(baseTime.AddDate(0, 1, 0)),
				TimeEnd:     timestamppb.New(baseTime.AddDate(0, 1, 0).Add(time.Hour)),
				Notify:      nil,
			},
			errString: "time is busy",
		},
	}

	for i, td := range testData {
		td := td
		s.T().Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			_, err := s.client.CreateEvent(ctx, td.rq)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)

			require.Equal(t, codes.InvalidArgument, st.Code())
			require.Contains(t, st.String(), td.errString)
		})
	}
}
