package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/logger"
)

type Server struct {
	server *http.Server
	logger logger.Logger
	events app.EventsUseCase
}

func New(logger logger.Logger, events app.EventsUseCase, addr string) *Server {
	s := newCalendarService(events, logger, time.Second*3, time.RFC3339)

	return &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: loggingMiddleware(createHandler(s), logger),
		},
		logger: logger,
		events: events,
	}
}

func (s *Server) Start() error {
	s.logger.Info("starting http server on " + s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server start: %w", err)
		}
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping http server")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	return nil
}

func createHandler(s *calendarAPI) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/", helloWorldHandler).Methods("GET")
	router.HandleFunc("/event", s.CreateHandler).Methods("POST")
	router.HandleFunc("/event/{id:[0-9]+}", s.GetByIDHandler).Methods("GET")
	router.HandleFunc("/event/{id:[0-9]+}", s.DeleteEventHandler).Methods("DELETE")
	router.HandleFunc("/event/{id:[0-9]+}", s.UpdateHandler).Methods("PUT")
	router.HandleFunc("/events/{period:day|week|month}", s.FindForPeriodHandler).Methods("GET")

	return router
}

func helloWorldHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, world"))
}
