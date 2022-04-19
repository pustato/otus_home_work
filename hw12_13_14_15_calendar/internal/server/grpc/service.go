package grpcserver

import (
	"context"
	"errors"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type calendarService struct {
	events app.EventsUseCase
	pb.UnimplementedCalendarServer
}

func newCalendarService(events app.EventsUseCase) *calendarService {
	return &calendarService{events: events}
}

func (s *calendarService) GetEvent(ctx context.Context, req *pb.EventRequest) (*pb.Event, error) {
	e, err := s.events.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, app.ErrEventIsNotExists) {
			return nil, status.Errorf(codes.NotFound, "event %d is not found", req.Id)
		}

		return nil, status.Errorf(codes.Internal, "grpc get event: %v", err.Error())
	}

	return eventToGrpc(e), nil
}

func (s *calendarService) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.EventResponse, error) {
	dto := app.CreateDTO{
		UserID:      req.UserId,
		Title:       req.Title,
		Description: req.Description,
		TimeStart:   req.TimeStart.AsTime(),
		TimeEnd:     req.TimeEnd.AsTime(),
		Notify:      req.Notify.AsDuration(),
	}

	id, err := s.events.Create(ctx, dto)
	if err != nil {
		var v *app.ValidationErrors
		if errors.As(err, &v) {
			return nil, status.Errorf(codes.InvalidArgument, "grpc create event validation error: %v", v.Error())
		}

		return nil, status.Errorf(codes.Internal, "grpc create event: %v", err.Error())
	}

	return &pb.EventResponse{
		Id: id,
	}, nil
}

func (s *calendarService) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*pb.EmptyResponse, error) {
	dto := app.UpdateDTO{
		Title:       req.Title,
		Description: req.Description,
		TimeStart:   req.TimeStart.AsTime(),
		TimeEnd:     req.TimeEnd.AsTime(),
		Notify:      req.Notify.AsDuration(),
	}

	err := s.events.Update(ctx, req.Id, dto)
	if err != nil {
		if errors.Is(err, app.ErrEventIsNotExists) {
			return nil, status.Errorf(codes.NotFound, "grpc create event: event %d is not exists", req.Id)
		}

		var v *app.ValidationErrors
		if errors.As(err, &v) {
			return nil, status.Errorf(codes.InvalidArgument, "grpc update event validation error: %v", v.Error())
		}

		return nil, status.Errorf(codes.Internal, "grpc update event: %v", err.Error())
	}

	return &pb.EmptyResponse{}, nil
}

func (s *calendarService) DeleteEvent(ctx context.Context, req *pb.EventRequest) (*pb.EmptyResponse, error) {
	if err := s.events.Delete(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "grpc delete event: %v", err.Error())
	}

	return &pb.EmptyResponse{}, nil
}

func (s *calendarService) FindForDay(ctx context.Context, req *pb.PeriodRequest) (*pb.EventCollection, error) {
	events, err := s.events.FindForDay(ctx, grpcPeriodToDto(req))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "grpc event get for day: %v", err.Error())
	}

	return eventsToGrpcCollection(events), nil
}

func (s *calendarService) FindForWeek(ctx context.Context, req *pb.PeriodRequest) (*pb.EventCollection, error) {
	events, err := s.events.FindForWeek(ctx, grpcPeriodToDto(req))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "grpc event get for week: %v", err.Error())
	}

	return eventsToGrpcCollection(events), nil
}

func (s *calendarService) FindForMonth(ctx context.Context, req *pb.PeriodRequest) (*pb.EventCollection, error) {
	events, err := s.events.FindForMonth(ctx, grpcPeriodToDto(req))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "grpc event get for month: %v", err.Error())
	}

	return eventsToGrpcCollection(events), nil
}

func grpcPeriodToDto(req *pb.PeriodRequest) app.FindByDateDTO {
	return app.FindByDateDTO{
		UserID: req.UserId,
		Date:   req.Date.AsTime(),
		Limit:  uint8(req.Limit),
		Offset: uint8(req.Offset),
	}
}

func eventToGrpc(e *storage.Event) *pb.Event {
	return &pb.Event{
		Id:          e.ID,
		UserId:      e.UserID,
		Title:       e.Title,
		Description: e.Description,
		TimeStart:   timestamppb.New(e.TimeStart),
		TimeEnd:     timestamppb.New(e.TimeEnd),
		NotifyAt: &pb.NullableNotificationTime{
			Valid: e.NotifyAt.Valid,
			Time:  timestamppb.New(e.NotifyAt.Time),
		},
		CreatedAt:        timestamppb.New(e.CreatedAt),
		UpdatedAt:        timestamppb.New(e.UpdatedAt),
		NotificationSent: e.NotificationSent,
	}
}

func eventsToGrpcCollection(events []*storage.Event) *pb.EventCollection {
	ev := make([]*pb.Event, 0, len(events))

	for _, e := range events {
		ev = append(ev, eventToGrpc(e))
	}

	return &pb.EventCollection{
		Events: ev,
	}
}
