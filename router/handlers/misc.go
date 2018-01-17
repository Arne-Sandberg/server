package handlers

import macaron "gopkg.in/macaron.v1"

// IndexHandler handles the / route, which is only GETtable.
// Note that this handler is not called if the user is not signed in. The /login handler
// will be called instaead.
// func (s Server) IndexHandler(c *macaron.Context) {
// 	user := c.Data["user"].(*models.User)
// 	files, err := s.filesystem.ListFilesForUser(user, ".")
// 	if err != nil {
// 		log.Error(0, "%v", err)
// 		c.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	c.HTML(200, "index", struct {
// 		Files       []os.FileInfo
// 		CurrentUser *models.User
// 	}{
// 		files,
// 		user,
// 	})
// }

func (s Server) NotFoundHandler(c *macaron.Context) {
	c.HTML(404, "notFound")
}
