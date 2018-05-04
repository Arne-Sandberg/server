package server

import (
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/server/httpRouter"
	"github.com/freecloudio/freecloud/server/grpcRouter"
)

func StartAll(httpPort int, grpcPort int, hostname string, virtualFS *fs.VirtualFilesystem) {
	httpRouter.Start(httpPort, hostname, virtualFS)
	grpcRouter.Start(grpcPort, hostname)
}

func StopAll() {
	httpRouter.Stop()
	grpcRouter.Stop()
}
