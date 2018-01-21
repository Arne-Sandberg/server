package handlers

import (
	"strconv"

	"github.com/riesinger/freecloud/auth"
	"github.com/riesinger/freecloud/models"
	macaron "gopkg.in/macaron.v1"
)

func (s Server) UserHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	user.Password = ""

	c.Data["response"] = user
}

func (s Server) UpdateUserHandler(c *macaron.Context) {

}

func (s Server) AdminUserHandler(c *macaron.Context) {
	userID, err := strconv.Atoi(c.Params("*"))
	if err != nil {
		c.Data["response"] = err
		return
	}
	user, err := auth.GetUserByID(userID)
	if err != nil {
		c.Data["response"] = err
		return
	}
	c.Data["response"] = user
}

func (s Server) AdminUpdateUserHandler(c *macaron.Context) {

}
