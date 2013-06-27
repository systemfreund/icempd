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

    ACK_ERROR_NOT_LIST = 1
    ACK_ERROR_ARG = 2
    ACK_ERROR_PASSWORD = 3
    ACK_ERROR_PERMISSION = 4
    ACK_ERROR_UNKNOWN = 5
    ACK_ERROR_NO_EXIST = 50
    ACK_ERROR_PLAYLIST_MAX = 51
    ACK_ERROR_SYSTEM = 52
    ACK_ERROR_PLAYLIST_LOAD = 53
    ACK_ERROR_UPDATE_ALREADY = 54
    ACK_ERROR_PLAYER_SYNC = 55
    ACK_ERROR_EXIST = 56
)

type Configuration struct {
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
	Conn net.Conn
	Dispatcher *MpdDispatcher
}

var nl = []byte {'\n'}

func (s *MpdSession) HandleEvents() {
	defer closeConn(s.Conn)

	// A new connection has been established, send welcome message
	s.Conn.Write([]byte(fmt.Sprintf("OK MPD %s\n", PROTOCOL_VERSION)))

	reader := bufio.NewScanner(s.Conn)
	for reader.Scan() {
		req := reader.Text()
		logger.Debug("%s --> %s", s.Conn.RemoteAddr(), req)
		resp, _ := s.Dispatcher.HandleRequest(req, 0)

		for _, line := range resp {
			logger.Debug("%s <-- %s", s.Conn.RemoteAddr(), line)	
			//s.Conn.Write([]byte(line))
		}
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
