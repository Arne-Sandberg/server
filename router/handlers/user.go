package handlers

import (
	"strconv"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	macaron "gopkg.in/macaron.v1"
)

var (
	allowedUserUpdates = map[string]bool{
		"FirstName": true,
		"LastName":  true,
		"AvatarURL": true,
		"Password":  true,
	}
	allowedAdminUpdates = map[string]bool{
		"FirstName": true,
		"LastName":  true,
		"AvatarURL": true,
		"Password":  true,
		"IsAdmin":   true,
		"Email":     true,
	}
)

func (s Server) UserHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)

	user.Password = ""
	c.Data["response"] = struct {
		Success bool         `json:"success"`
		User    *models.User `json:"user"`
	}{
		true,
		user,
	}
}

func (s Server) UpdateUserHandler(c *macaron.Context) {
	userID := c.Data["user"].(*models.User).ID
	updatedUser := c.Data["request"].(*models.User)

	updatedUser, err := auth.UpdateUser(userID, updatedUser, allowedUserUpdates)

	if err != nil {
		c.Data["response"] = err
	} else {
		c.Data["response"] = struct {
			Success bool         `json:"success"`
			User    *models.User `json:"user"`
		}{
			true,
			updatedUser,
		}
	}
}

func (s Server) AdminUserHandler(c *macaron.Context) {
	userID, err := strconv.Atoi(c.Params(":id"))
	if err != nil {
		c.Data["response"] = err
		return
	}
	user, err := auth.GetUserByID(userID)
	if err != nil {
		c.Data["response"] = err
		return
	}

	user.Password = ""
	c.Data["response"] = struct {
		Success bool         `json:"success"`
		User    *models.User `json:"user"`
	}{
		true,
		user,
	}
}

func (s Server) AdminUpdateUserHandler(c *macaron.Context) {
	userID, err := strconv.Atoi(c.Params(":id"))
	if err != nil {
		c.Data["response"] = err
		return
	}
	updatedUser := c.Data["request"].(*models.User)

	updatedUser, err = auth.UpdateUser(userID, updatedUser, allowedAdminUpdates)

	if err != nil {
		c.Data["response"] = err
	} else {
		c.Data["response"] = struct {
			Success bool         `json:"success"`
			User    *models.User `json:"user"`
		}{
			true,
			updatedUser,
		}
	}
}
