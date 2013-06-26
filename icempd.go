package main

import (
	"bufio"
	"fmt"
	"net"
	"github.com/op/go-logging"
	"code.google.com/p/gcfg"
)

var logger = logging.MustGetLogger("icempd")

const (
	PROTOCOL_ENCODING = "UTF-8"
	PROTOCOL_VERSION = "0.17.0"

	MSG_PASSWORD = "password"
)

type Configuration struct {
	Mpd struct {
		Listen string
		Password string
	}
}

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

func loadConfig() Configuration {
	result := Configuration{}
	err := gcfg.ReadFileInto(&result, "icempd.conf")
	if err != nil {
		logger.Fatalf("Configuration error: %s", err)
	}
	return result
}

func main() {
	config := loadConfig()
	listener, err := net.Listen("tcp", config.Mpd.Listen)
	checkError(err)

	for {
		logger.Debug("Wait")
		conn, err := listener.Accept()
		if err != nil {
			logger.Warning("Connection failed: %s", err.Error())
			continue
		}

		logger.Debug("New connection %s\n", conn.RemoteAddr())
		dispatcher := &MpdDispatcher{Config: config}

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
		logger.Fatalf("Fatal error: %s", err.Error())
	}
}
