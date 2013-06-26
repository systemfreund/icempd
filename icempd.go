package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("icempd")

const (
	PROTOCOL_ENCODING = "UTF-8"
	PROTOCOL_VERSION = "0.17.0"

	MSG_PASSWORD = "password"
)

type MpdSession struct {
	conn net.Conn
	dispatcher *MpdDispatcher
}

func (s *MpdSession) HandleEvents() {
	defer closeConn(s.conn)

	// A new connection has been established, send welcome message
	s.conn.Write([]byte(fmt.Sprintf("OK MPD %s\n", PROTOCOL_VERSION)))

	reader := bufio.NewScanner(s.conn)
	for reader.Scan() {
		req := Request(reader.Text())
		logger.Debug("< %s\n", req)
		s.dispatcher.HandleRequest(&req, 0)
	}
}

func main() {
	service := ":6600"
	listener, err := net.Listen("tcp", service)
	checkError(err)

	for {
		logger.Debug("Wait")
		conn, err := listener.Accept()
		if err != nil {
			logger.Warning("Connection failed: %s", err.Error())
			continue
		}

		logger.Debug("New connection %s\n", conn.RemoteAddr())
		dispatcher := new(MpdDispatcher)

		session := MpdSession{conn, dispatcher}
		go session.HandleEvents()
	}
}

func closeConn(conn net.Conn) {
	logger.Debug("Close connection %s\n", conn.RemoteAddr())
	defer conn.Close()
}

func checkError(err error) {
	if err != nil {
		logger.Fatal("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}
