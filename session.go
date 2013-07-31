package main

import (
	"github.com/op/go-logging"
)

type Session struct {
	Core
	Connection

	Authenticated        bool
	commandListReceiving bool
	commandListOk        bool
	commandList          []string
	commandListIndex     int

	subscriptions  []string
	events         []string
	preventTimeout bool

	*logging.Logger
}

func NewSession(core Core, conn Connection) (s Session) {
	s = Session {
		Core: core,
		Connection: conn,
		Logger: logging.MustGetLogger(LOGGER_NAME),
	}

	return
}

func (s *Session) isCurrentlyIdle() bool {
	return nil != s.subscriptions
}

func (s *Session) addSubscription(sub string) {
	s.Debug("add subscription: %s", sub)
	// TODO dont add dupes
	s.subscriptions = append(s.subscriptions, sub)
}

func (s *Session) clearSubscriptions() {
	s.Debug("clear subscriptions")
	s.subscriptions = nil
}

func (s *Session) addEvent(subsystem string) {
	// TODO add event
	s.Debug("add event: %s", subsystem)
}

func (s *Session) clearEvents() {
	s.Debug("clear events")
	s.events = nil
}

func (s *Session) getActiveEvents() []string {
	// TODO return intersection of events and subscriptions
	return []string{}
}
