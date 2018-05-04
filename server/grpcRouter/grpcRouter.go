package grpcRouter

import (
	"fmt"
	"net"

	"github.com/freecloudio/freecloud/models"
	"google.golang.org/grpc"
	log "gopkg.in/clog.v1"
)

var grpcServer grpc.Server

func Start(port int, hostname string) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		log.Fatal(0, "grpc: failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	models.RegisterAuthServiceServer(grpcServer, NewAuthService())
	models.RegisterUserServiceServer(grpcServer, NewUserService())
	models.RegisterFilesServiceServer(grpcServer, NewFilesService())
	models.RegisterSystemServiceServer(grpcServer, NewSystemService())

	// Start server in a goroutine so the method exits and all interrupts can be handled correctly
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			log.Fatal(0, "Server error: %v", err)
		}
	}()
}

// Stop shutdowns the currently running server
func Stop() {
	grpcServer.GracefulStop()
}
