package main

import (
	"fmt"
	"io"
	"os"
	"github.com/op/go-logging"
	"code.google.com/p/gcfg"
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
		Path string
		DbPath string
	}

	Logging struct {
		Level int
	}

	Mpd struct {
		Listen string
		Password string
	}
}

type MpdAckError struct {
	Code int
	Index int
	Command string
	Message string
}

func (e MpdAckError) Error() string {
	return e.Message
}

func (e *MpdAckError) AckString() string {
	return fmt.Sprintf("ACK [%d@%d] {%s} %s", e.Code, e.Index, e.Command, e.Message)
}

type MpdSession struct {
	Id string
	Config Configuration

	Authenticated bool
	commandListReceiving bool
	commandListOk bool
	commandList []string
	commandListIndex int

	io.ReadWriteCloser
	*logging.Logger
}

func NewMpdSession(id string, conn io.ReadWriteCloser, config Configuration) (s MpdSession) {
	s = MpdSession{
		Id: id,
		Config: config,
		ReadWriteCloser: conn,
		Logger: logging.MustGetLogger(LOGGER_NAME),
	}

	s.Notice("New session %s", s.Id)

	return
}

func (s *MpdSession) Close() {
	s.Info("Close session %s", s.Id)
	s.ReadWriteCloser.Close()
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
	<- server.Stop
}
