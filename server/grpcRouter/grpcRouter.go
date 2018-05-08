package grpcRouter

import (
	"fmt"
	"github.com/freecloudio/freecloud/models"
	"github.com/freecloudio/freecloud/fs"
	"google.golang.org/grpc"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	log "gopkg.in/clog.v1"
	"net/http"
	"context"
)

var httpServer http.Server

func Start(port int, hostname string, vfs *fs.VirtualFilesystem) {
	grpcServer := grpc.NewServer()
	models.RegisterAuthServiceServer(grpcServer, NewAuthService(vfs))
	models.RegisterUserServiceServer(grpcServer, NewUserService())
	models.RegisterFilesServiceServer(grpcServer, NewFilesService(vfs))
	models.RegisterSystemServiceServer(grpcServer, NewSystemService())

	wrappedGrpc := grpcweb.WrapServer(grpcServer)
	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedGrpc.ServeHTTP(resp, req)
	}

	httpServer = http.Server{
		Addr:    fmt.Sprintf("%s:%d", hostname, port),
		Handler: http.HandlerFunc(handler),
	}

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
	err := httpServer.Shutdown(context.Background())
	if err != nil {
		log.Error(0, "Error shutting down grpcServer: %v", err)
	}
}
