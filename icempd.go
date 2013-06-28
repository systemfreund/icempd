package main

import (
	"bufio"
	"fmt"
	"net"
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
	Conn net.Conn
	Config Configuration
	Dispatcher MpdDispatcher
}

func NewMpdSession(conn net.Conn, config Configuration) MpdSession {
	result := MpdSession{
		Conn: conn,
		Config: config,
	}

	result.Dispatcher = MpdDispatcher{ Session: &result }

	return result
}

var nl = []byte {'\n'}

func (s *MpdSession) HandleEvents() {
	defer closeConn(s.Conn)

	// A new connection has been established, send welcome message
	s.Conn.Write([]byte(fmt.Sprintf("OK MPD %s\n", PROTOCOL_VERSION)))

	reader := bufio.NewScanner(s.Conn)
	for reader.Scan() {
		req := reader.Text()
		logger.Info("%s --> %s", s.Conn.RemoteAddr(), req)
		resp, _ := s.Dispatcher.HandleRequest(req, 0)

		for _, line := range resp {
			logger.Info("%s <-- %s", s.Conn.RemoteAddr(), line)	
			s.Conn.Write(append([]byte(line), '\n'))
		}
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

func init() {
	config = loadConfig()
	logging.SetLevel(logging.Level(config.Logging.Level), LOGGER_NAME)
}

func main() {
	listener, err := net.Listen("tcp", config.Mpd.Listen)
	if err != nil {
		fmt.Printf("Fatal error: %s", err.Error())
		os.Exit(2)
	}

	NewLibrary(config.Library.Path)

	logger.Notice("Listen at %s", config.Mpd.Listen)
	for {
		logger.Debug("Wait")
		conn, err := listener.Accept()
		if err != nil {
			logger.Warning("Connection failed: %s", err.Error())
			continue
		}

		logger.Info("New connection %s\n", conn.RemoteAddr())

		session := NewMpdSession(conn, config)
		go session.HandleEvents()
	}
}

func closeConn(conn net.Conn) {
	logger.Info("Close connection %s\n", conn.RemoteAddr())
	defer conn.Close()
}
