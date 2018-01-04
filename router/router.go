package router

import (
	"fmt"
	"net/http"

	"github.com/riesinger/freecloud/auth"
	"github.com/riesinger/freecloud/config"
	"github.com/riesinger/freecloud/fs"
	log "gopkg.in/clog.v1"
	macaron "gopkg.in/macaron.v1"
)

type server struct {
	filesystem fs.Filesystem
}

var (
	s server
)

func Start(port int, hostname string, filesys fs.Filesystem, credProvider auth.CredentialsProvider) {
	if config.GetBool("http.ssl") {
		log.Warn("SSL is not implemented yet, falling back to HTTP")
	}
	log.Info("Starting router at http://%s:%d", hostname, port)
	s = server{
		filesystem: filesys,
	}

	m := macaron.New()
	m.Use(s.Logging())
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("public", macaron.StaticOptions{SkipLogging: true}))
	m.Use(macaron.Static("node_modules/uikit/dist", macaron.StaticOptions{SkipLogging: true}))
	m.Use(macaron.Renderer())

	m.Post("/upload", s.OnlyUsers, s.FileUpload)
	m.Get("/", s.OnlyUsers, s.IndexHandler)
	m.Get("/signup", s.OnlyAnonymous, s.SignupPageHandler)
	m.Post("/signup", s.SignupHandler)
	m.Get("/login", s.OnlyAnonymous, s.LoginPageHandler)
	m.Post("/login", s.LoginHandler)

	m.NotFound(s.NotFoundHandler)

	log.Fatal(0, "%v", http.ListenAndServe(fmt.Sprintf("%s:%d", hostname, port), m))
}
