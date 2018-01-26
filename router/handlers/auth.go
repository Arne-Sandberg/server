package handlers

import (
	"net/http"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"
)

func (s Server) LoginHandler(c *macaron.Context) {
	userIntf, ok := c.Data["request"]
	if !ok {
		log.Error(0, "%v", ErrNoRequestData)
		c.Data["response"] = ErrNoRequestData
		return
	}

	user := userIntf.(*models.User)

	session, err := auth.NewSession(user.Email, user.Password)
	if err == auth.ErrInvalidCredentials {
		log.Info("Invalid credentials for user %s", user.Email)
		c.Data["response"] = models.APIError{Code: http.StatusUnauthorized, Message: "Wrong credentials or account does not exist"}
		return
	}
	if err != nil {
		// TODO: Catch the "not found" error and also return StatusUnauthorized here
		log.Error(0, "Failed to get user %s: %v", user.Email, err)
		c.Data["response"] = err
		return
	}
	c.Data["response"] = struct {
		Success bool   `json:"success"`
		Token   string `json:"token"`
	}{
		Success: true,
		Token:   session.GetTokenString(),
	}
}

// LogoutHandler deletes the active user session and signs out the user.
func (s Server) LogoutHandler(c *macaron.Context) {
	session := c.Data["session"].(models.Session)
	err := auth.RemoveSession(session)
	if err != nil {
		log.Error(0, "Failed to remove session during logout: %v", err)
		c.Data["response"] = models.APIError{Code: http.StatusInternalServerError, Message: "Failed to delete session"}
		return
	}
	c.Data["response"] = models.SuccessResponse
}

// SignupHandler handles the /signup route, when a POST request is made to it.
// It creates a new user and returns a session and user cookie.
func (s Server) SignupHandler(c *macaron.Context) {
	userIntf, ok := c.Data["request"]
	if !ok {
		log.Error(0, "%v", ErrNoRequestData)
		c.Data["response"] = ErrNoRequestData
		return
	}

	user := userIntf.(*models.User)

	log.Trace("Signing up user: %s %s with email %s", user.FirstName, user.LastName, user.Email)
	session, err := auth.NewUser(user)
	if err == auth.ErrInvalidUserData || err == auth.ErrUserAlreadyExists {
		c.Data["response"] = models.APIError{Code: http.StatusBadRequest, Message: err.Error()}
		return
	} else if err != nil {
		c.Data["response"] = err
		return
	}
	c.Data["response"] = struct {
		Success bool   `json:"success"`
		Token   string `json:"token"`
	}{
		Success: true,
		Token:   session.GetTokenString(),
	}
}
