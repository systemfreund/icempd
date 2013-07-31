package main

import (
	"fmt"
	"github.com/op/go-logging"
	"net"
	"os"
	"io"
)

type Server struct {
	Connections chan Connection
	Stop     chan bool

	*logging.Logger
}

type Connection struct {
	Id     string

	io.ReadWriteCloser
	logging.Logger
}

func NewServer(config Configuration) (s Server) {
	listener, err := net.Listen("tcp", config.Mpd.Listen)
	if err != nil {
		fmt.Printf("Fatal error: %s", err.Error())
		os.Exit(2)
	}

	s = Server{
		make(chan Connection),
		make(chan bool),
		logging.MustGetLogger(LOGGER_NAME),
	}
	s.Notice("Listen at %s", config.Mpd.Listen)

	go func() {
		for {
			s.Debug("Wait")
			conn, err := listener.Accept()
			if err != nil {
				s.Warning("Connection failed: %s", err.Error())
				continue
			}

			s.Debug("New Connection from %v\n", conn.RemoteAddr())
			s.Connections <- newConnection(conn.RemoteAddr().String(), conn)
		}
	}()

	return
}

func newConnection(id string, conn io.ReadWriteCloser) (s Connection) {
	s = Connection{id, conn, *logging.MustGetLogger(LOGGER_NAME)}
	s.Notice("New Connection %s", s.Id)
	return
}

func (s *Connection) Close() {
	s.Info("Close Connection %s", s.Id)
	s.ReadWriteCloser.Close()
}
