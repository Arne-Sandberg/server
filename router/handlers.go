package router

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/riesinger/freecloud/auth"

	"github.com/riesinger/freecloud/config"
	"github.com/riesinger/freecloud/models"
	log "gopkg.in/clog.v1"
	macaron "gopkg.in/macaron.v1"
)

const oneGigabyte = 1024 * 1024 * 1024 * 1024

func (s server) FileUpload(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	// Parse the multipart form in the request
	err := c.Req.ParseMultipartForm(config.GetInt64("http.upload_limit") * oneGigabyte)
	if err != nil {
		log.Error(0, "File upload failed: %v", err)
		c.WriteHeader(http.StatusInternalServerError)
		return
	}

	multiform := c.Req.MultipartForm

	// Get the *fileheaders
	files, ok := multiform.File["files"]
	if !ok {
		log.Error(0, "No 'files' form field, aborting file upload")
		c.WriteHeader(http.StatusBadRequest)
		return
	}
	for i := range files {
		// For each fileheader, get a handle to the actual file
		file, err := files[i].Open()
		if err != nil {
			log.Error(0, "Could not open file: %v", err)
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Create the destination file making sure the path is writeable.
		dst, err := s.filesystem.NewFileHandleForUser(user, files[i].Filename)
		if err != nil {
			log.Error(0, "Could not open file for writing: %v", err)
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			log.Error(0, "Could not copy the file: %v", err)
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	c.WriteHeader(http.StatusCreated)
	//c.HTML(http.StatusCreated, "files/upload", "Upload successful!")
}

func (s server) SignupPageHandler(c *macaron.Context) {
	c.HTML(200, "auth/signup")
}

// SignupHandler handles the /signup route, when a POST request is made to it.
// It creates a new user and returns a session and user cookie.
func (s server) SignupHandler(c *macaron.Context) {
	if c.Req.Request.Body == nil {
		log.Warn("No user data received while signing up")
		c.WriteHeader(http.StatusBadRequest)
		return
	}
	defer c.Req.Request.Body.Close()
	// Deserialize user
	log.Trace("Deserializing user")

	var user models.User

	err := json.NewDecoder(c.Req.Request.Body).Decode(&user)
	if err != nil {
		log.Error(0, "Could not decode user data: %v", err)
		c.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Trace("Signing up user: %s %s with email %s", user.FirstName, user.LastName, user.Email)
	session, err := auth.NewUser(&user)
	if err != nil {
		c.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.SetCookie(config.GetString("auth.session_cookie"), session.GetCookieString())
	c.WriteHeader(http.StatusOK)
}

// IndexHandler handles the / route, which is only GETtable.
// Note that this handler is not called if the user is not signed in. The /login handler
// will be called instaead.
func (s server) IndexHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	files, err := s.filesystem.ListFilesForUser(user, ".")
	if err != nil {
		log.Error(0, "%v", err)
		c.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.HTML(200, "index", struct {
		Files       []os.FileInfo
		CurrentUser *models.User
	}{
		files,
		user,
	})
}

func (s server) LoginPageHandler(c *macaron.Context) {
	c.HTML(http.StatusOK, "auth/login", nil)
}

func (s server) LoginHandler(c *macaron.Context) {
	type jsonData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if c.Req.Request.Body == nil {
		log.Warn("No user data received while signing in")
		c.WriteHeader(http.StatusBadRequest)
		return
	}
	defer c.Req.Request.Body.Close()
	// Deserialize user
	log.Trace("Deserializing login data")
	var data jsonData
	err := json.NewDecoder(c.Req.Request.Body).Decode(&data)
	if err != nil {
		log.Error(0, "Could not decode login data: %v", err)
		c.WriteHeader(http.StatusInternalServerError)
		return
	}

	session, err := auth.NewSession(data.Email, data.Password)
	if err == auth.ErrInvalidCredentials {
		log.Info("Invalid credentials for user %s", data.Email)
		c.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Error(0, "Failed to get user %s: %v", data.Email, err)
		c.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.SetCookie(config.GetString("auth.session_cookie"), session.GetCookieString())
	c.WriteHeader(http.StatusOK)
}

func (s server) LogoutHandler(c *macaron.Context) {
	session := c.Data["session"].(models.Session)
	err := auth.RemoveSession(session)
	if err != nil {
		log.Error(0, "Failed to remove session during logout: %v", err)
		// Don't set an InternalServerError header because the logout should run normally on the client
	}

	c.SetCookie(config.GetString("auth.session_cookie"), "", -1) // Set a MaxAge of -1 to delete the cookie
	c.WriteHeader(http.StatusOK)
}

// func (server) ListUsersHandler(c *macaron.Context) {
// 	users := db.
// 	c.HTML(200, "listUsers", struct{
// 		Users []*model.User
// 	}{
// 		users,
// 	})
// }

func (server) NotFoundHandler(c *macaron.Context) {
	c.HTML(404, "notFound")
}
