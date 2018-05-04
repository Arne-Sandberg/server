package httpRouter

import (
	"fmt"
	"net/http"

	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/fs"
	"gopkg.in/macaron.v1"
	log "gopkg.in/clog.v1"
)

type ServerHandler struct {
	filesystem *fs.VirtualFilesystem
}

var (
	s          ServerHandler
	httpServer http.Server
)

// Start starts the router with the given settings
func Start(port int, hostname string, virtualFS *fs.VirtualFilesystem) {
	if config.GetBool("http.ssl") {
		log.Warn("SSL is not implemented yet, falling back to HTTP")
	}
	log.Info("Starting router at http://%s:%d", hostname, port)
	s = ServerHandler{filesystem: virtualFS}

	m := macaron.New()
	m.Use(Logging())
	m.Use(macaron.Recovery())

	// Up- and Download of files
	m.Get("/download/*", OnlyUsers, ResolvePath, s.DownloadHandler)
	m.Post("/upload/*", OnlyUsers, ResolvePath, s.UploadHandler, JSONEncoder)

	m.Use(macaron.Static("client/dist", macaron.StaticOptions{SkipLogging: true}))
	m.NotFound(s.NotFoundHandler)

	httpServer = http.Server{Addr: fmt.Sprintf("%s:%d", hostname, port), Handler: m}

	// Start server in a goroutine so the method exits and all interrupts can be handled correctly
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
	s = ServerHandler{}
}
