package handlers

import (
	"strconv"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	macaron "gopkg.in/macaron.v1"
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
	c.Data["response"] = user
}

func (s Server) AdminUpdateUserHandler(c *macaron.Context) {

}
