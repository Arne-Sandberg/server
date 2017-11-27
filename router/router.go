package router

import (
	"fmt"
	"net/http"

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
	log.Info("Starting router at %s:%d", hostname, port)
	m := macaron.Classic()
	m.Use(macaron.Renderer())

	s = server{}
	filesystem = filesys
	m.Get("/ping", s.PingHandler)
	m.Get("/upload", s.FileUpload)
	m.Post("/upload", s.FileUpload)

	log.Fatal(0, "%v", http.ListenAndServe(fmt.Sprintf("%s:%d", hostname, port), m))
}
