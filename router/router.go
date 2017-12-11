package router

import (
	"fmt"
	"net/http"
	"os"

	"github.com/riesinger/freecloud/config"
	"github.com/riesinger/freecloud/models"
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
	m := macaron.Classic()
	m.Use(macaron.Renderer())

	s = server{}
	filesystem = filesys
	m.Get("/ping", s.PingHandler)
	m.Get("/upload", s.FileUpload)
	m.Post("/upload", s.FileUpload)
	m.Get("/list", s.FileList)

	m.Get("/", func(c *macaron.Context) {
		files, err := filesystem.ListFiles(".")
		if err != nil {
			c.Error(http.StatusInternalServerError, err.Error())
			return
		}
		c.HTML(200, "index", struct{ 
			Files []os.FileInfo
			CurrentUser models.User
			}{
				files,
				models.User{SignedIn: false},
			})
	})

	log.Fatal(0, "%v", http.ListenAndServe(fmt.Sprintf("%s:%d", hostname, port), m))
}
