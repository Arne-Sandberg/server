package router

import (
	"net/http"
	"strconv"
	"time"

	"github.com/riesinger/freecloud/auth"
	"github.com/riesinger/freecloud/config"

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

	if session, user := c.GetCookie(config.GetString("auth.session_cookie")), c.GetCookie(config.GetString("auth.user_cookie")); session == "" || user == "" {
		c.Redirect("/login", http.StatusFound)
		return
	} else {
		// convert the user cookie to a user id
		userID, err := strconv.ParseInt(user, 10, 0)
		if err != nil {
			log.Error(0, "Could not convert userID %s to an integer", user, err)
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
		valid, err := auth.ValidateSession(int(userID), auth.Session(session))
		if err != nil {
			log.Error(0, "Could not validate session: %v", err)
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !valid {
			log.Warn("Invalid session")
			c.Redirect("/login", http.StatusFound)
			return
		}

		// If the session is valid, fill the context's user data
		user, err := s.credentialsProvider.GetUserByID(int(userID))
		if err != nil {
			log.Warn("Filling user data in middleware failed: %v", err)
		}
		user.SignedIn = true
		c.Data["user"] = user
	}

}
