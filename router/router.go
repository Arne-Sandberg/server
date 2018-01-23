package router

import (
	"fmt"
	"net/http"

	"github.com/freecloudio/freecloud/router/handlers"

	"github.com/freecloudio/freecloud/models"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/fs"
	log "gopkg.in/clog.v1"
	macaron "gopkg.in/macaron.v1"
)

var (
	s handlers.Server
)

// Start starts the router with the given settings
func Start(port int, hostname string, filesys fs.Filesystem, credProvider auth.CredentialsProvider) {
	if config.GetBool("http.ssl") {
		log.Warn("SSL is not implemented yet, falling back to HTTP")
	}
	log.Info("Starting router at http://%s:%d", hostname, port)
	s = handlers.NewServer(filesys)

	m := macaron.New()
	m.Use(Logging())
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("public", macaron.StaticOptions{SkipLogging: true}))
	m.Use(macaron.Static("node_modules/uikit/dist", macaron.StaticOptions{SkipLogging: true}))
	m.Use(macaron.Renderer())

	m.Group("/api/v1", func() {
		m.Group("/auth", func() {
			m.Post("/signup", OnlyAnonymous, JSONDecoder(&models.User{}), s.SignupHandler, JSONEncoder)
			m.Post("/login", OnlyAnonymous, JSONDecoder(&models.User{}), s.LoginHandler, JSONEncoder)
			m.Post("/logout", OnlyUsers, s.LogoutHandler, JSONEncoder)
		})

		m.Get("/user/me", OnlyUsers, s.UserHandler, JSONEncoder)
		//m.Patch("/user/me", OnlyUsers, JSONDecoder(&models.User{}), s.UpdateUserHandler, JSONEncoder)
		m.Get("/user/*", OnlyAdmins, s.AdminUserHandler, JSONEncoder)
		//m.Patch("/user/*", OnlyAdmins, JSONDecoder(&models.User{}), s.AdminUpdateUserHandler, JSONEncoder)

		m.Post("/files", OnlyUsers, s.UploadHandler, JSONEncoder)
		m.Get("/directory/*", OnlyUsers, s.GetDirectoryHandler, JSONEncoder)
	})

	// TODO: implement handlers for the client's static files

	// TODO: We need to differenciate between 404s coming from the API vs ones coming from the user typing a bad URL.
	// We then should either send a simple 404 or redirect to sth like /#/404
	m.NotFound(s.NotFoundHandler)

	log.Fatal(0, "%v", http.ListenAndServe(fmt.Sprintf("%s:%d", hostname, port), m))
}
