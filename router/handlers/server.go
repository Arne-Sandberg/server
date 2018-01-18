package handlers

import (
	"errors"

	"github.com/riesinger/freecloud/fs"
)

var ErrNoRequestData = errors.New("Expected request data from JSONDecoder, but got none")

type Server struct {
	filesystem fs.Filesystem
}

func NewServer(filesystem fs.Filesystem) Server {
	return Server{filesystem: filesystem}
}
