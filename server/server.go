package server

import (
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/server/httpRouter"
	"github.com/freecloudio/freecloud/server/grpcRouter"
)

func StartAll(httpPort, webGrpcPort, natGrpcPort int, hostname string, virtualFS *fs.VirtualFilesystem, startNatGrpc bool) {
	httpRouter.Start(httpPort, hostname, virtualFS)
	grpcRouter.Start(webGrpcPort, natGrpcPort, hostname, virtualFS, startNatGrpc)
}

func StopAll() {
	httpRouter.Stop()
	grpcRouter.Stop()
}
