package grpcRouter

import (
	"google.golang.org/grpc"
	"time"
	"context"
	log "gopkg.in/clog.v1"
)

func loggingInterceptor(ctx context.Context, req interface{},	info *grpc.UnaryServerInfo,	handler grpc.UnaryHandler) (interface{}, error) {
	startTime := time.Now()
	log.Info("Started %s", info.FullMethod)

	h, err := handler(ctx, req)

	if err != nil {
		log.Error(0, "Finished %s in %v with error: %v", info.FullMethod, time.Since(startTime), err)
	} else {
		log.Info("Finished %s in %v", info.FullMethod, time.Since(startTime))
	}

	return h, err
}
