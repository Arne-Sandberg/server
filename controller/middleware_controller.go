package controller

import (
	"net/http"
	"strings"
	"time"

	errors "github.com/go-openapi/errors"
	log "gopkg.in/clog.v1"

	"github.com/freecloudio/server/manager"
	"github.com/freecloudio/server/models"
)

func FileServerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/swagger.json") {
			next.ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir("./client/build")).ServeHTTP(w, r)
		}
	})
}

func ValidateToken(token string, scopes []string) (principal *models.Principal, err error) {
	principal = &models.Principal{Token: &models.Token{Token: token}}

	if len(scopes) > 0 {
		var session *models.Session
		session, err = models.ParseSessionString(token)
		if err != nil {
			return nil, errors.New(http.StatusUnauthorized, "Token could not be parsed")
		}

		valid := manager.GetAuthManager().ValidateSession(session)
		if !valid {
			return nil, errors.New(http.StatusUnauthorized, "No valid session")
		}

		principal.User, err = manager.GetAuthManager().GetUserByID(session.UserID)
		if err != nil {
			return nil, errors.New(http.StatusInternalServerError, err.Error())
		}

		if isUserScope(scopes) || (isAdminScope(scopes) && principal.User.IsAdmin) {
			return
		} else {
			return nil, errors.New(http.StatusForbidden, "Insufficient privileges")
		}
	}

	return
}

// StatusRecordingResponseWriter is a wrapper around a http.ResponseWriter, which allows us to
// read the status code that has been returned. This is useful for logging.
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

// LoggingMiddleware logs incoming requests and their responses
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

func isUserScope(scopes []string) bool {
	userScope := false
	for _, scope := range scopes {
		if scope == "user" {
			userScope = true
			break
		}
	}

	return userScope
}

func isAdminScope(scopes []string) bool {
	adminScope := false
	for _, scope := range scopes {
		if scope == "admin" {
			adminScope = true
			break
		}
	}

	return adminScope
}
