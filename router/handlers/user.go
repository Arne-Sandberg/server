package handlers

import (
	"strconv"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/models"
	"github.com/go-restit/lzjson"
	macaron "gopkg.in/macaron.v1"
)

var (
	allowedUserUpdates = []string{
		"firstName",
		"lastName",
		"avatarURL",
		"password",
	}
	allowedAdminUpdates = []string{
		"firstName",
		"lastName",
		"avatarURL",
		"password",
		"isAdmin",
		"email",
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

func (s Server) UserListHandler(c *macaron.Context) {
	users, err := auth.GetAllUsers()
	if err != nil {
		c.Data["response"] = err
		return
	}
	c.Data["response"] = struct {
		Success bool           `json:"success"`
		Users   []*models.User `json:"users"`
	}{
		true,
		users,
	}
}

func (s Server) UpdateUserHandler(c *macaron.Context) {
	userID := c.Data["user"].(*models.User).ID
	userUpdateJSON := c.Data["request"].(lzjson.Node)

	updatedUser, err := auth.UpdateUser(userID, fillUserUpdates(userUpdateJSON, false))

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
	userUpdateJSON := c.Data["request"].(lzjson.Node)

	updatedUser, err := auth.UpdateUser(userID, fillUserUpdates(userUpdateJSON, true))

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

func fillUserUpdates(userUpdateJSON lzjson.Node, admin bool) (updates map[string]interface{}) {
	updates = make(map[string]interface{})

	var allowedUpdates *[]string
	if admin {
		allowedUpdates = &allowedAdminUpdates
	} else {
		allowedUpdates = &allowedUserUpdates
	}

	var temp interface{}
	for _, identifier := range *allowedUpdates {
		value := userUpdateJSON.Get(identifier)
		if err := value.ParseError(); err != nil {
			continue
		}

		value.Unmarshal(&temp)
		updates[identifier] = temp
	}

	return
}
