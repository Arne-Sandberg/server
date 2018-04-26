package router

import (
	"fmt"
	"net/http"

	"github.com/freecloudio/freecloud/router/handlers"

	"github.com/freecloudio/freecloud/models"
	apiModels "github.com/freecloudio/freecloud/models/api"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/fs"
	log "gopkg.in/clog.v1"
	macaron "gopkg.in/macaron.v1"
)

var (
	s          handlers.Server
	httpServer http.Server
)

// Start starts the router with the given settings
func Start(port int, hostname string, virtualFS *fs.VirtualFilesystem, credProvider auth.CredentialsProvider) {
	if config.GetBool("http.ssl") {
		log.Warn("SSL is not implemented yet, falling back to HTTP")
	}
	log.Info("Starting router at http://%s:%d", hostname, port)
	s = handlers.NewServer(virtualFS)

	m := macaron.New()
	m.Use(Logging())
	m.Use(macaron.Recovery())
	m.Use(macaron.Renderer())

	m.Group("/api/v1", func() {
		// Auth: Includes logging in/out and signup
		m.Group("/auth", func() {
			m.Post("/signup", JSONDecoder(&models.User{}), s.SignupHandler, JSONEncoder)
			m.Post("/login", JSONDecoder(&models.User{}), s.LoginHandler, JSONEncoder)
			m.Post("/logout", OnlyUsers, s.LogoutHandler, JSONEncoder)
		})

		// User: Includes getting and editing your user or as admin also for other users
		m.Get("/users", OnlyUsers, s.UserListHandler, JSONEncoder)
		m.Get("/user/me", OnlyUsers, s.UserHandler, JSONEncoder)
		m.Patch("/user/me", OnlyUsers, GeneralJSONDecoder, s.UpdateUserHandler, JSONEncoder)
		m.Get("/user/byID/:id", OnlyAdmins, s.AdminUserHandler, JSONEncoder)
		m.Patch("/user/byID/:id", OnlyAdmins, GeneralJSONDecoder, s.AdminUpdateUserHandler, JSONEncoder)
		m.Delete("/user/me", OnlyUsers, s.DeleteUserHandler, JSONEncoder)
		m.Delete("/user/byID/:id", OnlyAdmins, s.AdminDeleteUserHandler, JSONEncoder)

		// Data: Up- and Download of files, creation and modifying files and directories
		m.Get("/download/*", OnlyUsers, ResolvePath, s.DownloadHandler, JSONEncoder)
		m.Post("/zip", OnlyUsers, JSONDecoder(&apiModels.ZipRequest{}), s.ZipHandler, JSONEncoder)

		m.Post("/upload/*", OnlyUsers, ResolvePath, s.UploadHandler, JSONEncoder)
		m.Get("/path/*", OnlyUsers, ResolvePath, s.FileInfoHandler, JSONEncoder)
		m.Post("/path/*", OnlyUsers, ResolvePath, JSONDecoder(&models.FileInfo{}), s.CreateFileHandler, JSONEncoder)
		m.Patch("/path/*", OnlyUsers, ResolvePath, GeneralJSONDecoder, s.UpdateFileHandler, JSONEncoder)
		m.Delete("/path/*", OnlyUsers, ResolvePath, s.FileDeleteHandler, JSONEncoder)

		m.Post("/share/*", OnlyUsers, ResolvePath, JSONDecoder(&apiModels.ShareRequest{}), s.ShareHandler, JSONEncoder)

		m.Get("/search/*", OnlyUsers, ResolvePath, s.SearchHandler, JSONEncoder)
		m.Get("/starred", OnlyUsers, s.StarredFilesInfoHandler, JSONEncoder)
		m.Get("/shared", OnlyUsers, s.SharedFilesInfoHandler, JSONEncoder)

		m.Get("/rescan/me", OnlyUsers, s.RescanHandler, JSONEncoder)
		m.Get("/rescan/byID/:id", OnlyAdmins, s.AdminRescanHandler, JSONEncoder)

		m.Get("/stats", OnlyAdmins, s.StatsHandler, JSONEncoder)
	})

	m.Use(macaron.Static("client/dist", macaron.StaticOptions{SkipLogging: true}))
	m.NotFound(s.NotFoundHandler)

	httpServer = http.Server{Addr: fmt.Sprintf("%s:%d", hostname, port), Handler: m}

	// Start server in a goroutine so the method exits and all interrupts can be handled correclty
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			log.Fatal(0, "Server error: %v", err)
		}
	}()
}

// Stop shutdowns the currently running server
func Stop() {
	if httpServer.Addr == "" {
		return
	}

	if err := httpServer.Shutdown(nil); err != nil {
		log.Fatal(0, "Error shutting down server: %v", err)
		return
	}

	httpServer = http.Server{}
	s = handlers.Server{}
}
