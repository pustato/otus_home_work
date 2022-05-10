package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
)

const internalError = "internal server error"

type response struct {
	Error *string     `json:"error"`
	Data  interface{} `json:"data"`
}

type createEventRequest struct {
	UserID      int64  `json:"userId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	TimeStart   string `json:"timeStart"`
	TimeEnd     string `json:"timeEnd"`
	Notify      string `json:"notify"`
}

type updateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	TimeStart   string `json:"timeStart"`
	TimeEnd     string `json:"timeEnd"`
	Notify      string `json:"notify"`
}

type createEventResponse struct {
	ID int64 `json:"id"`
}

type eventResponse struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"userId"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	TimeStart   string  `json:"timeStart"`
	TimeEnd     string  `json:"timeEnd"`
	NotifyAt    *string `json:"notifyAt"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type calendarAPI struct {
	timeLayout string
	events     app.EventsUseCase
	log        logger.Logger
	timeout    time.Duration
}

func newCalendarService(
	events app.EventsUseCase,
	log logger.Logger,
	timeout time.Duration,
	timeLayout string,
) *calendarAPI {
	return &calendarAPI{
		events:     events,
		log:        log,
		timeout:    timeout,
		timeLayout: timeLayout,
	}
}

func (s *calendarAPI) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.logErrorf("http event get: id is not int: %s", err.Error())
		s.writeErrorResponse(w, "invalid id", http.StatusBadRequest)
		return
	}

	e, err := s.events.GetByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, app.ErrEventIsNotExists) {
			s.writeErrorResponse(w, "not found", http.StatusNotFound)
			return
		}

		s.logErrorf("http event get: events use case: %s", err.Error())
		s.writeErrorResponse(w, internalError, http.StatusInternalServerError)
		return
	}

	s.writeResponse(w, &response{
		nil,
		s.storageEventToResponse(e),
	}, http.StatusOK)
}

func (s *calendarAPI) CreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	rq := &createEventRequest{}
	if err := json.NewDecoder(r.Body).Decode(rq); err != nil {
		s.logErrorf("http event create: decode request: %s", err.Error())
		s.writeErrorResponse(w, "malformed json", http.StatusBadRequest)
		return
	}

	dto, err := s.createRequestToDTO(rq)
	if err != nil {
		s.logErrorf("http event create: %s", err.Error())
		s.writeErrorResponse(w, "invalid request", http.StatusBadRequest)
		return
	}

	id, err := s.events.Create(ctx, *dto)
	if err != nil {
		var v *app.ValidationErrors
		if errors.As(err, &v) {
			s.writeErrorResponse(w, v.Error(), http.StatusUnprocessableEntity)
			return
		}

		s.logErrorf("http event create: events use case: %s", err.Error())
		s.writeErrorResponse(w, internalError, http.StatusInternalServerError)
		return
	}

	s.writeResponse(w, &response{nil, createEventResponse{id}}, http.StatusCreated)
}

func (s *calendarAPI) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.logErrorf("http event update: id is not int: %s", err.Error())
		s.writeErrorResponse(w, "invalid id", http.StatusBadRequest)
		return
	}

	rq := &updateEventRequest{}
	if err := json.NewDecoder(r.Body).Decode(rq); err != nil {
		s.logErrorf("http event update: decode request: %s", err.Error())
		s.writeErrorResponse(w, "malformed json", http.StatusBadRequest)
		return
	}

	dto, err := s.updateRequestToDTO(rq)
	if err != nil {
		s.logErrorf("http event update: %s", err.Error())
		s.writeErrorResponse(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := s.events.Update(ctx, int64(id), *dto); err != nil {
		var v *app.ValidationErrors
		if errors.As(err, &v) {
			s.writeErrorResponse(w, v.Error(), http.StatusUnprocessableEntity)
			return
		}

		s.logErrorf("http event update: events use case: %s", err.Error())
		s.writeErrorResponse(w, internalError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *calendarAPI) DeleteEventHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.logErrorf("http event delete: id is not int: %s", err.Error())
		s.writeErrorResponse(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := s.events.Delete(ctx, int64(id)); err != nil {
		s.logErrorf("http event delete: events use case: %s", err.Error())
		s.writeErrorResponse(w, internalError, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *calendarAPI) FindForPeriodHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()
	var err error

	dto := app.FindByDateDTO{}
	q := r.URL.Query()

	u, ok := q["userId"]
	if !ok {
		s.writeErrorResponse(w, "`userId` is required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(u[0])
	if err != nil {
		s.writeErrorResponse(w, "`userId` must be numeric", http.StatusBadRequest)
		return
	}
	dto.UserID = int64(userID)

	d, ok := q["date"]
	if !ok {
		s.writeErrorResponse(w, "`date` is required", http.StatusBadRequest)
		return
	}
	dto.Date, err = time.Parse(s.timeLayout, d[0])
	if err != nil {
		s.writeErrorResponse(w, "`date` must has layout "+s.timeLayout, http.StatusBadRequest)
		return
	}

	l, ok := q["limit"]
	if !ok {
		dto.Limit = 50
	} else {
		limit, err := strconv.Atoi(l[0])
		if err != nil {
			s.writeErrorResponse(w, "`limit` must be numeric", http.StatusBadRequest)
			return
		}
		dto.Limit = uint8(limit)
	}

	o, ok := q["offset"]
	if !ok {
		dto.Offset = 0
	} else {
		offset, err := strconv.Atoi(o[0])
		if err != nil {
			s.writeErrorResponse(w, "`offset` must be numeric", http.StatusBadRequest)
			return
		}
		dto.Offset = uint8(offset)
	}

	vars := mux.Vars(r)
	period := vars["period"]

	var events []*storage.Event
	switch period {
	case "day":
		events, err = s.events.FindForDay(ctx, dto)
	case "week":
		events, err = s.events.FindForWeek(ctx, dto)
	case "month":
		events, err = s.events.FindForMonth(ctx, dto)
	default:
		s.writeErrorResponse(w, "unknown period", http.StatusNotFound)
		return
	}

	if err != nil {
		s.logErrorf("http find for %s: event use case: %v", period, err.Error())
		s.writeErrorResponse(w, internalError, http.StatusBadRequest)
		return
	}

	rspEvents := make([]*eventResponse, 0, len(events))
	for _, e := range events {
		rspEvents = append(rspEvents, s.storageEventToResponse(e))
	}
	s.writeResponse(w, &response{nil, rspEvents}, http.StatusOK)
}

func (s *calendarAPI) writeResponse(w http.ResponseWriter, rsp *response, statusCode int) {
	w.Header().Add("Content-Type", "application/json")
	body, err := json.Marshal(rsp)
	if err != nil {
		s.logErrorf("marshal response: %s", err.Error())
		return
	}

	if _, err := w.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logErrorf("writer response: %s", err.Error())
		return
	}
	w.WriteHeader(statusCode)
}

func (s *calendarAPI) writeErrorResponse(w http.ResponseWriter, msg string, statusCode int) {
	rsp := response{&msg, nil}
	s.writeResponse(w, &rsp, statusCode)
}

func (s *calendarAPI) logErrorf(format string, a ...interface{}) {
	s.log.Error(fmt.Sprintf(format, a...),
		"context", "http",
	)
}

func (s *calendarAPI) storageEventToResponse(e *storage.Event) *eventResponse {
	var notify *string
	if e.NotifyAt.Valid {
		t := e.NotifyAt.Time.Format(s.timeLayout)
		notify = &t
	}

	return &eventResponse{
		ID:          e.ID,
		UserID:      e.UserID,
		Title:       e.Title,
		Description: e.Description,
		TimeStart:   e.TimeStart.Format(s.timeLayout),
		TimeEnd:     e.TimeEnd.Format(s.timeLayout),
		NotifyAt:    notify,
		CreatedAt:   e.CreatedAt.Format(s.timeLayout),
		UpdatedAt:   e.CreatedAt.Format(s.timeLayout),
	}
}

func (s *calendarAPI) createRequestToDTO(r *createEventRequest) (*app.CreateDTO, error) {
	ts, err := time.Parse(s.timeLayout, r.TimeStart)
	if err != nil {
		return nil, fmt.Errorf("parse time start: %w", err)
	}

	te, err := time.Parse(s.timeLayout, r.TimeEnd)
	if err != nil {
		return nil, fmt.Errorf("parse time end: %w", err)
	}

	notify, err := time.ParseDuration(r.Notify)
	if err != nil {
		return nil, fmt.Errorf("parse notify: %w", err)
	}

	return &app.CreateDTO{
		UserID:      r.UserID,
		Title:       r.Title,
		Description: r.Description,
		TimeStart:   ts,
		TimeEnd:     te,
		Notify:      notify,
	}, nil
}

func (s *calendarAPI) updateRequestToDTO(r *updateEventRequest) (*app.UpdateDTO, error) {
	ts, err := time.Parse(s.timeLayout, r.TimeStart)
	if err != nil {
		return nil, fmt.Errorf("parse time start: %w", err)
	}

	te, err := time.Parse(s.timeLayout, r.TimeEnd)
	if err != nil {
		return nil, fmt.Errorf("parse time end: %w", err)
	}

	notify, err := time.ParseDuration(r.Notify)
	if err != nil {
		return nil, fmt.Errorf("parse notify: %w", err)
	}

	return &app.UpdateDTO{
		Title:       r.Title,
		Description: r.Description,
		TimeStart:   ts,
		TimeEnd:     te,
		Notify:      notify,
	}, nil
}
