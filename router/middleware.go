package router

import (
	"net/http"
	"time"

	"github.com/riesinger/freecloud/auth"
	"github.com/riesinger/freecloud/config"
	"github.com/riesinger/freecloud/models"

	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"
)

func (s server) Logging() macaron.Handler {
	return func(c *macaron.Context) {
		startTime := time.Now()
		log.Info("Started %s %s for %s", c.Req.Method, c.Req.RequestURI, c.RemoteAddr())

		rw := c.Resp.(macaron.ResponseWriter)
		c.Next()
		elapsed := time.Since(startTime)

		switch rw.Status() {
		case 401, 403, 404:
			log.Warn("Finished %s %s in %v, %d %s", c.Req.Method, c.Req.RequestURI, elapsed, rw.Status(), http.StatusText(rw.Status()))
		case 500:
			log.Error(0, "Finished %s %s in %v, %d %s", c.Req.Method, c.Req.RequestURI, elapsed, rw.Status(), http.StatusText(rw.Status()))
		default:
			log.Info("Finished %s %s in %v, %d %s", c.Req.Method, c.Req.RequestURI, elapsed, rw.Status(), http.StatusText(rw.Status()))
		}

	}
}

func (s server) IsUser(c *macaron.Context) {

	if sessionStr := c.GetCookie(config.GetString("auth.session_cookie")); sessionStr == "" {
		c.Redirect("/login", http.StatusFound)
		return
	} else {
		// convert the user cookie to a user id
		session, err := models.ParseSessionCookieString(sessionStr)
		if err != nil {
			log.Error(0, "Could not parse session token: %v", err)
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
		valid := auth.ValidateSession(session)

		if !valid {
			log.Warn("Invalid session")
			c.Redirect("/login", http.StatusFound)
			return
		}

		// If the session is valid, fill the context's user data
		user, err := s.credentialsProvider.GetUserByID(session.UID)
		if err != nil {
			log.Warn("Filling user data in middleware failed: %v", err)
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
		user.SignedIn = true
		c.Data["user"] = user
		c.Data["session"] = session
	}

}
