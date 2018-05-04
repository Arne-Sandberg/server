package server

import (
	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/server/httpRouter"
)

func StartServers(port int, hostname string, virtualFS *fs.VirtualFilesystem, credProvider auth.CredentialsProvider) {
	httpRouter.Start(port, hostname, virtualFS, credProvider)
}
