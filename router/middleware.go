package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/models"
	apiModels "github.com/freecloudio/freecloud/models/api"
	"github.com/go-restit/lzjson"

	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"
)

// Logging is a middleware logging ever request with URL, IP, duration and status
func Logging() macaron.Handler {
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

// OnlyAdmins is the same as OnlyUsers, however we have an additional check to only allow admins to pass through.
func OnlyAdmins(c *macaron.Context) {
	OnlyUsers(c)
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

// OnlyUsers only allows users to pass through, sending a http.StatusUnauthorized to non-users.
// If a valid session/user pair is encountered, this fills the context's "session" and "user" fields.
func OnlyUsers(c *macaron.Context) {
	// First, check the Authorization header, as this is the preferred method
	// for authentication.
	sessionStr := c.Req.Header.Get("Authorization")
	if sessionStr != "" {
		validateSessionAndFillUserData(sessionStr, c)
		return
	}
	sessionStr = c.GetCookie(config.GetString("auth.session_cookie"))
	if sessionStr != "" {
		validateSessionAndFillUserData(sessionStr, c)
		return
	}
	c.WriteHeader(http.StatusUnauthorized)
	return
}

func validateSessionAndFillUserData(token string, c *macaron.Context) {
	session, err := models.ParseSessionTokenString(token)
	// This probably also means the session is invalid, so redirect time it is!
	if err != nil {
		log.Error(0, "Could not parse session token: %v", err)
		c.WriteHeader(http.StatusUnauthorized)
		return
	}
	valid := auth.ValidateSession(session)
	if !valid {
		log.Warn("Invalid session")
		c.WriteHeader(http.StatusUnauthorized)
		return
	}

	// If the session is valid, fill the context's user data
	user, err := auth.GetUserByID(session.UID)
	if err != nil {
		log.Warn("Filling user data in middleware failed: %v", err)
		c.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.Data["user"] = user
	c.Data["session"] = session
}

// OnlyAnonymous only allows non-users to pass through, otherwise we'll write a http.StatusForbidden
func OnlyAnonymous(c *macaron.Context) {
	if c.GetCookie(config.GetString("auth.session_cookie")) == "" && c.Header().Get("Authorization") == "" {
		// We were successfully identified as nobody ;)
		return
	}
	c.WriteHeader(http.StatusForbidden)
}

// GeneralJSONDecoder unmarshals the context's body into a json object of the lzjson package
// This object can hold up any json structure and the further handlers need to unpack it themself
func GeneralJSONDecoder(c *macaron.Context) {
	json := lzjson.Decode(c.Req.Request.Body)
	defer c.Req.Request.Body.Close()

	c.Data["request"] = json
}

// JSONDecoder unmarshals the context's body into a variable of type "to" into the context's "request" data field.
// If this fails, we exit by sending an internal server error code.
// Note that the "to" variable will be overwritten.
func JSONDecoder(to interface{}) macaron.Handler {
	return func(c *macaron.Context) {
		if c.Req.Request.Body == nil {
			log.Warn("Got no JSON request payload, expected a %T", to)
			c.WriteHeader(http.StatusBadRequest)
			return
		}
		defer c.Req.Request.Body.Close()

		decoded := reflect.New(reflect.Indirect(reflect.ValueOf(to)).Type()).Interface()
		err := json.NewDecoder(c.Req.Request.Body).Decode(decoded)
		if err != nil {
			log.Error(0, "Could not unmarshal payload into a %T: %v", to, err)
			c.WriteHeader(http.StatusBadRequest)
			return
		}

		c.Data["request"] = decoded
	}
}

// JSONEncoder marshals the payload inside of c.Data["resposne"] as JSON and sends it to the client.
// There are some notables cases though:
// If "response" is a models.APIError, we'll marshal and send it with the APIError's Code.
// If "response" is a regular error, we'll send the same message as if it were an APIError, but with code 500 instead.
// Otherwise, we'll send a response with a 200 OK and the JSON-ified payload.
func JSONEncoder(c *macaron.Context) {
	resp, ok := c.Data["response"]
	if !ok {
		// TODO: Decide whether this is the correct logging level
		log.Error(0, "No payload to marshal")
		c.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch res := resp.(type) {
	case apiModels.Error:
		if res.Code != 0 {
			c.JSON(res.Code, res)
			return
		}
		c.JSON(http.StatusInternalServerError, res)
		return
	case error:
		c.JSON(http.StatusInternalServerError, struct {
			Message string `json:"message"`
		}{
			res.Error(),
		})
		return
	default:
		c.JSON(http.StatusOK, res)
		return
	}

}

// ResolvePath resolves the * parameter of a url and creates a user independent path of it
func ResolvePath(c *macaron.Context) {
	path, err := url.PathUnescape(c.Params("*"))
	if err != nil {
		c.Data["response"] = fmt.Errorf("Invalid path format")
	}
	c.Data["path"] = path
}
