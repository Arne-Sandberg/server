package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/clog.v1"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/db"
	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/router"
	"github.com/freecloudio/freecloud/stats"
)

var (
	VERSION = "0.1.0"
)

func main() {
	stats.Init(VERSION, time.Now().UTC())
	config.Init()
	setupLogger()

	filesystem, err := fs.NewDiskFilesystem(config.GetString("fs.base_directory"), config.GetString("fs.tmp_folder_name")) // TODO: Remove temp folder name from dfs and move completely to vfs
	if err != nil {
		os.Exit(3)
	}

	database, err := db.NewStormDB(config.GetString("db.name"))
	if err != nil {
		clog.Fatal(0, "Database setup failed, bailing out!")
		os.Exit(1)
	}

	auth.Init(database, database)

	virtualFS, err := fs.NewVirtualFilesystem(filesystem, database, config.GetString("fs.tmp_folder_name"))

	router.Start(config.GetInt("http.port"), config.GetString("http.host"), virtualFS, database)
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
