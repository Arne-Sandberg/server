package router

import (
	"fmt"
	"net/http"

	"github.com/riesinger/freecloud/config"
	"github.com/riesinger/freecloud/fs"
	log "gopkg.in/clog.v1"
	macaron "gopkg.in/macaron.v1"
)

type server struct{}

var (
	s          server
	filesystem fs.Filesystem
)

func Start(port int, hostname string, filesys fs.Filesystem) {
	if config.GetBool("http.ssl") {
		log.Warn("SSL is not implemented yet, falling back to HTTP")
	}
	log.Info("Starting router at http://%s:%d", hostname, port)
	s = server{}

	m := macaron.New()
	m.Use(s.Logging())
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("public", macaron.StaticOptions{SkipLogging: true}))
	m.Use(macaron.Renderer())

	filesystem = filesys
	m.Post("/upload", s.FileUpload)
	m.Get("/", s.IndexHandler)
	m.Get("/signup", s.SignupPageHandler)
	m.Post("/signup", s.SignupHandler)

	log.Fatal(0, "%v", http.ListenAndServe(fmt.Sprintf("%s:%d", hostname, port), m))
}
