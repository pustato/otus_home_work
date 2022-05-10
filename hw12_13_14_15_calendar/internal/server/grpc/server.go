package grpcserver

import (
	"fmt"
	"net"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"google.golang.org/grpc"
)

type Server struct {
	addr   string
	logger logger.Logger
	events app.EventsUseCase
	server *grpc.Server
}

func New(logger logger.Logger, events app.EventsUseCase, addr string) *Server {
	return &Server{
		addr:   addr,
		logger: logger,
		events: events,
	}
}

func (s *Server) Start() error {
	s.logger.Info("listening on " + s.addr)
	lsn, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listening grpc on %s: %w", s.addr, err)
	}

	s.server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			unaryLoggingInterceptor(s.logger),
		),
	)
	pb.RegisterCalendarServer(s.server, newCalendarService(s.events))

	s.logger.Info("starting grpc server")
	if err := s.server.Serve(lsn); err != nil {
		return fmt.Errorf("starting grpc server: %w", err)
	}

	return nil
}

func (s *Server) Stop() {
	s.server.Stop()
}
