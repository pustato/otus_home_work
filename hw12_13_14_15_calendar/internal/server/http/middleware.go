package httpserver

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/logger"
)

const timeLayout = "[02/Jan/2006:15:04:05 -0700]"

type responseWriterDecorator struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterDecorator) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriterDecorator {
	return &responseWriterDecorator{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func loggingMiddleware(next http.Handler, log logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		decoratedWriter := wrapResponseWriter(w)
		next.ServeHTTP(decoratedWriter, r)

		msg := strings.Join([]string{
			r.RemoteAddr,
			start.Format(timeLayout),
			r.Method,
			r.URL.Path,
			r.Proto,
			strconv.Itoa(decoratedWriter.statusCode),
			time.Since(start).String(),
			r.UserAgent(),
		}, " ")

		log.Info(msg,
			"type", "access",
			"context", "http",
		)
	})
}
