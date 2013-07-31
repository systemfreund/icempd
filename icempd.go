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
	// library := NewLibrary(config.Library.Path)
	// NewSqliteTagDb(config.Library.DbPath, library.TuneChannel)
	playlist := Playlist{}
	addFakeTune(playlist)
	server := NewServer(config)
	core := Core{config, playlist}
	NewDispatcher(server.Sessions, core)
	<-server.Stop
}

func addFakeTune(p Playlist) {
	t := Tune{
		Uri:    "file:///Users/yildiz/tmp/music/Quantic - Mishaps Happening (2004)/03 sound of everything.mp3",
		Title:  "Mishaps Happening",
		Artist: "Quantic",
		Album:  "Mishaps Happening",
		Genre:  "Downbeat",
		Year:   2004,
		Track:  3,
		Length: 242,
	}

	pe := PlaylistEntry{
		id:   0,
		tune: t,
	}

	p.add(pe)
}
