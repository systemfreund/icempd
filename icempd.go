package main

import (
	"bufio"
	"net"
	//"net/textproto"
	"os"
	"fmt"
)

const (
	PROTOCOL_ENCODING = "UTF-8"
	PROTOCOL_VERSION = "0.17.0"
)

type MpdSession struct {
	conn net.Conn
}

func (s *MpdSession) HandleEvents() {
	defer closeConn(s.conn)

	// A new connection has been established, send welcome message
	s.conn.Write([]byte(fmt.Sprintf("OK MPD %s\n", PROTOCOL_VERSION)))

	reader := bufio.NewScanner(s.conn)
	for reader.Scan() {
		fmt.Fprintf(os.Stdout, "Got: '%s'\n", reader.Text())
	}
}

func main() {
	service := ":6600"
	listener, err := net.Listen("tcp", service)
	checkError(err)

	for {
		fmt.Fprintf(os.Stdout, "Listening for new connection\n")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't accept connection: %s", err.Error())
			continue
		}

		fmt.Fprintf(os.Stdout, "New connection from %s\n", conn.RemoteAddr())
		session := MpdSession{conn}
		go session.HandleEvents()
	}
}

func closeConn(conn net.Conn) {
	fmt.Fprintf(os.Stdout, "Closing connection from %s\n", conn.RemoteAddr())
	defer conn.Close()
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}
