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

	subscriptions  []string
	events         []string
	preventTimeout bool

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

func (s *MpdSession) isCurrentlyIdle() bool {
	return nil != s.subscriptions
}

func (s *MpdSession) addSubscription(sub string) {
	s.Debug("add subscription: %s", sub)
	// TODO dont add dupes
	s.subscriptions = append(s.subscriptions, sub)
}

func (s *MpdSession) clearSubscriptions() {
	s.Debug("clear subscriptions")
	s.subscriptions = nil
}

func (s *MpdSession) addEvent(subsystem string) {
	// TODO add event
	s.Debug("add event: %s", subsystem)
}

func (s *MpdSession) clearEvents() {
	s.Debug("clear events")
	s.events = nil
}

func (s *MpdSession) getActiveEvents() []string {
	// TODO return intersection of events and subscriptions
	return []string{}
}
