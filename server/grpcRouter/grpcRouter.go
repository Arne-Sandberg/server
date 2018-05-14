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
	"net"
)

var webGrpcServer *http.Server
var startNatFlag bool
var grpcServer *grpc.Server

func Start(webPort, natPort int, hostname string, vfs *fs.VirtualFilesystem, startNat bool) {
	grpcServer = grpc.NewServer()
	models.RegisterAuthServiceServer(grpcServer, NewAuthService(vfs))
	models.RegisterUserServiceServer(grpcServer, NewUserService())
	models.RegisterFilesServiceServer(grpcServer, NewFilesService(vfs))
	models.RegisterSystemServiceServer(grpcServer, NewSystemService())

	wrappedWebGrpc := grpcweb.WrapServer(grpcServer)
	webHandler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedWebGrpc.ServeHTTP(resp, req)
	}

	webGrpcServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", hostname, webPort),
		Handler: http.HandlerFunc(webHandler),
	}

	// Start server in a goroutine so the method exits and all interrupts can be handled correctly
	go func() {
		err := webGrpcServer.ListenAndServe()
		if err != nil {
			log.Fatal(0, "Server error: %v", err)
		}
	}()

	startNatFlag = startNat
	if startNatFlag {
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", hostname, natPort))
		if err != nil {
			log.Fatal(0, "grpc: failed to listen: %v", err)
			return
		}

		// Start server in a goroutine so the method exits and all interrupts can be handled correctly
		go func() {
			err := grpcServer.Serve(lis)
			if err != nil {
				log.Fatal(0, "Server error: %v", err)
			}
		}()
	}
}

// Stop shutdowns the currently running server
func Stop() {
	err := webGrpcServer.Shutdown(context.Background())
	if err != nil {
		log.Error(0, "Error shutting down grpcServer: %v", err)
	}

	if startNatFlag {
		grpcServer.GracefulStop()
	}
}
