package utils

import (
	"fmt"
	"os"

	clog "gopkg.in/clog.v1"
)

func SetupLogger() {
	err := clog.New(clog.CONSOLE, clog.ConsoleConfig{
		Level: clog.TRACE,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize logging: %v", err)
		os.Exit(2)
	}
}

func CloseLogger() {
	clog.Shutdown()
}
