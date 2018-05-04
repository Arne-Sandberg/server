package grpcRouter

import (
	"fmt"
	"log"
	"net"

	"github.com/freecloudio/freecloud/models"
	"google.golang.org/grpc"
)

var grpcServer grpc.Server

func Start(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	models.RegisterAuthServiceServer(grpcServer, NewAuthService())
	models.RegisterUserServiceServer(grpcServer, NewUserService())
	models.RegisterFilesServiceServer(grpcServer, NewFilesService())
	models.RegisterSystemServiceServer(grpcServer, NewSystemService())

	// Start server in a goroutine so the method exits and all interrupts can be handled correclty
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			log.Fatal(0, "Server error: %v", err)
		}
	}()
}

// Stop shutdowns the currently running server
func Stop() {
	grpcServer.Stop()
	grpcServer = grpc.Server{}
}
