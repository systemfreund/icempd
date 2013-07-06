package main

import (
	"fmt"
	"net"
	"os"
	"github.com/op/go-logging"
)

type Server struct {
	Sessions chan MpdSession
	Stop chan bool

	*logging.Logger
}

func NewServer(config Configuration) (s Server) {
	listener, err := net.Listen("tcp", config.Mpd.Listen)
	if err != nil {
		fmt.Printf("Fatal error: %s", err.Error())
		os.Exit(2)
	}

	s = Server{
		make(chan MpdSession), 
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
			s.Sessions <- NewMpdSession(conn.RemoteAddr().String(), conn, config)
		}
	}()

	return
}