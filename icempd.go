package main

import (
	"code.google.com/p/gcfg"
	"fmt"
	"github.com/op/go-logging"
	"os"
)

const (
	LOGGER_NAME = "icempd"
)

var (
	logger = logging.MustGetLogger(LOGGER_NAME)
	config Configuration
)

type Configuration struct {
	Library struct {
		Path   string
		DbPath string
	}

	Logging struct {
		Level int
	}

	Mpd struct {
		Listen   string
		Password string
	}
}

func loadConfig() Configuration {
	result := Configuration{}
	err := gcfg.ReadFileInto(&result, "icempd.conf")
	if err != nil {
		fmt.Printf("Configuration error: %s", err)
		os.Exit(1)
	}
	return result
}

func main() {
	config = loadConfig()
	logging.SetLevel(logging.Level(config.Logging.Level), LOGGER_NAME)
	library := NewLibrary(config.Library.Path)
	NewSqliteTagDb(config.Library.DbPath, library.TuneChannel)
	server := NewServer(config)
	NewDispatcher(config, server.Sessions)
	<-server.Stop
}
