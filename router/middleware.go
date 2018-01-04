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

func (s server) OnlyAdmins(c *macaron.Context) {
	s.OnlyUsers(c)
	userRaw, ok := c.Data["user"]
	if !ok {
		return
	}
	user := userRaw.(*models.User)
	if user.IsAdmin {
		return
	}
	c.WriteHeader(http.StatusForbidden)
}

func (s server) OnlyUsers(c *macaron.Context) {
	if sessionStr := c.GetCookie(config.GetString("auth.session_cookie")); sessionStr == "" {
		c.Redirect("/login", http.StatusFound)
		return
	} else {
		session, err := models.ParseSessionCookieString(sessionStr)
		// This probably also means the session is invalid, so redirect time it is!
		if err != nil {
			log.Error(0, "Could not parse session token: %v", err)
			c.SetCookie(config.GetString("auth.session_cookie"), "", -1)
			c.Redirect("/login", http.StatusFound)
			return
		}
		valid := auth.ValidateSession(session)
		if !valid {
			log.Warn("Invalid session")
			c.SetCookie(config.GetString("auth.session_cookie"), "", -1)
			c.Redirect("/login", http.StatusFound)
			return
		}

		// If the session is valid, fill the context's user data
		user, err := auth.GetUserByID(session.UID)
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

func (s server) OnlyAnonymous(c *macaron.Context) {
	if sessionStr := c.GetCookie(config.GetString("auth.session_cookie")); sessionStr == "" {
		// We were successfully identified as nobody ;)
		return
	} else {
		c.Redirect("/", http.StatusFound)
		return
	}
}
