package main

import (
	"github.com/op/go-logging"
	"io"
)

type MpdSession struct {
	Id     string
	Config Configuration

	Authenticated        bool
	commandListReceiving bool
	commandListOk        bool
	commandList          []string
	commandListIndex     int

	io.ReadWriteCloser
	*logging.Logger
}

func NewMpdSession(id string, conn io.ReadWriteCloser, config Configuration) (s MpdSession) {
	s = MpdSession{
		Id:              id,
		Config:          config,
		ReadWriteCloser: conn,
		Logger:          logging.MustGetLogger(LOGGER_NAME),
	}

	s.Notice("New session %s", s.Id)

	return
}

func (s *MpdSession) Close() {
	s.Info("Close session %s", s.Id)
	s.ReadWriteCloser.Close()
}
