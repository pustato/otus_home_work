package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/application"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/logger"
)

type Server struct {
	server *http.Server
	logger logger.Logger
	events application.EventsUseCase
}

func New(logger logger.Logger, events application.EventsUseCase, addr string) *Server {
	return &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: loggingMiddleware(createHandler(), logger),
		},
		logger: logger,
		events: events,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting http server on " + s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil {
		return fmt.Errorf("http server start: %w", err)
	}

	return s.server.Shutdown(ctx)
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping http server")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	return nil
}

func createHandler() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/", helloWorldHandler).Methods("GET")

	return router
}

func helloWorldHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, world"))
}
