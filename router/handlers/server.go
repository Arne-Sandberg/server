package handlers

import (
	"errors"

	"github.com/freecloudio/freecloud/fs"
)

var ErrNoRequestData = errors.New("Expected request data from JSONDecoder, but got none")

type Server struct {
	filesystem *fs.VirtualFilesystem
}

func NewServer(filesystem *fs.VirtualFilesystem) Server {
	return Server{filesystem: filesystem}
}
