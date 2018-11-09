package utils

import (
	"net/http"
	"time"

	log "gopkg.in/clog.v1"
)

type StatusRecordingResponseWriter struct {
	status int
	http.ResponseWriter
}

func NewStatusRecordingResponseWriter(res http.ResponseWriter) *StatusRecordingResponseWriter {
	return &StatusRecordingResponseWriter{200, res}
}

func (w *StatusRecordingResponseWriter) Status() int {
	return w.status
}

func (w *StatusRecordingResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *StatusRecordingResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

func (w *StatusRecordingResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Trace("%s request to %s from %s", r.Method, r.URL.Path, r.Host)
		startTime := time.Now()
		srw := NewStatusRecordingResponseWriter(w)
		next.ServeHTTP(srw, r)
		endTime := time.Now()
		log.Trace("%s to %s took %v, response is %d %s", r.Method, r.URL.Path, endTime.Sub(startTime), srw.Status(), http.StatusText(srw.Status()))
	})
}
