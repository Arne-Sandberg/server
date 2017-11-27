package main

import (
	"fmt"
	"os"

	"gopkg.in/clog.v1"

	"github.com/riesinger/freecloud/config"
	"github.com/riesinger/freecloud/fs"
	"github.com/riesinger/freecloud/router"
)

func main() {
	config.Init()
	setupLogger()
	filesystem, err := fs.NewDiskFilesystem(config.GetString("fs.base_directory"))
	if err != nil {
		os.Exit(3)
	}
	router.Start(config.GetInt("http.port"), config.GetString("http.host"), filesystem)
}

func setupLogger() {
	err := clog.New(clog.CONSOLE, clog.ConsoleConfig{
		Level: clog.TRACE,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize logging: %v", err)
		os.Exit(2)
	}
}
