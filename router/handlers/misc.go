package handlers

import (
	"net/http"
	"strings"

	macaron "gopkg.in/macaron.v1"
)

func (s Server) NotFoundHandler(c *macaron.Context) {
	if strings.Contains(c.Req.RequestURI, "api/v") {
		c.JSON(http.StatusNotFound, struct {
			Message string `json:"message"`
		}{
			"404: not found",
		})
		return
	}

	c.Redirect("/#/404")
}
