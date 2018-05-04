package server

import (
	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/server/httpRouter"
	"github.com/freecloudio/freecloud/server/grpcRouter"
)

func StartAll(httpPort int, grpcPort int, hostname string, virtualFS *fs.VirtualFilesystem, credProvider auth.CredentialsProvider) {
	httpRouter.Start(httpPort, hostname, virtualFS, credProvider)
	grpcRouter.Start(grpcPort)
}

func StopAll() {
	httpRouter.Stop()
	grpcRouter.Stop()
}
