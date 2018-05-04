package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/clog.v1"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/db"
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/stats"
	"github.com/freecloudio/freecloud/server"
)

var (
	VERSION = "0.1.0"
)

func main() {
	stats.Init(VERSION, time.Now().UTC())
	config.Init()
	setupLogger()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exitChan := make(chan int)
	go func() {
		for {
			s := <-signalChan
			switch s {
			// kill -SIGHUP XXXX
			case syscall.SIGHUP:
				fallthrough
			// kill -SIGINT XXXX or Ctrl+c
			case syscall.SIGINT:
				fallthrough
			// kill -SIGTERM XXXX
			case syscall.SIGTERM:
				fallthrough
			// kill -SIGQUIT XXXX
			case syscall.SIGQUIT:
				clog.Info("Stopping freecloud due to a signal: %v", s)
				exitChan <- 0

			default:
				clog.Info("Stopping freecloud due to an unknown signal: %v", s)
				exitChan <- 1
			}
		}
	}()

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

	server.StartAll(config.GetInt("http.port"), config.GetInt("grpc.port"), config.GetString("net.host"), virtualFS)

	code := <-exitChan

	server.StopAll()
	virtualFS.Close()
	auth.Close()
	database.Close()
	filesystem.Close()
	clog.Info("freecloud stopped.")
	closeLogger()
	os.Exit(code)
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

func closeLogger() {
	clog.Shutdown()
}
