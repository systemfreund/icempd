package main

import (
	"fmt"
	"regexp"
)

const (
	PROTOCOL_ENCODING = "UTF-8"
	PROTOCOL_VERSION  = "0.17.0"

	ACK_ERROR_NOT_LIST       = 1
	ACK_ERROR_ARG            = 2
	ACK_ERROR_PASSWORD       = 3
	ACK_ERROR_PERMISSION     = 4
	ACK_ERROR_UNKNOWN        = 5
	ACK_ERROR_NO_EXIST       = 50
	ACK_ERROR_PLAYLIST_MAX   = 51
	ACK_ERROR_SYSTEM         = 52
	ACK_ERROR_PLAYLIST_LOAD  = 53
	ACK_ERROR_UPDATE_ALREADY = 54
	ACK_ERROR_PLAYER_SYNC    = 55
	ACK_ERROR_EXIST          = 56
)

type MpdAckError struct {
	Code    int
	Index   int
	Command string
	Message string
}

func (e MpdAckError) Error() string {
	return e.Message
}

func (e *MpdAckError) AckString() string {
	return fmt.Sprintf("ACK [%d@%d] {%s} %s", e.Code, e.Index, e.Command, e.Message)
}

type CommandHandlerFunc func(session *Session, params map[string]string) ([]string, error)

type MpdCommand struct {
	AuthRequired bool
	Handler      CommandHandlerFunc
	Pattern      *regexp.Regexp
}

var (
	MPD_COMMANDS map[string]MpdCommand
	SUBSYSTEMS   []string
)

func init() {
	MPD_COMMANDS = map[string]MpdCommand{
		// Connection
		"close":    MpdCommand{false, closeMpdConn, regexp.MustCompile("^close$")},
		"ping":     MpdCommand{false, ping, regexp.MustCompile("^ping$")},
		"password": MpdCommand{false, password, regexp.MustCompile("^password \"(?P<password>[^\"]+)\"$")},

		// Status
		"status": MpdCommand{true, mpdStatus, regexp.MustCompile("^status$")},
		"idle":   MpdCommand{true, setIdle, regexp.MustCompile("^idle( (?P<subsystems>.+))?$")},
		"noidle": MpdCommand{true, setNoIdle, regexp.MustCompile("^noidle$")},

		// Playlist
		"playlistinfo": MpdCommand{true, getPlaylistInfo, regexp.MustCompile("^playlistinfo$")},
	}

	SUBSYSTEMS = []string{
		"database", "mixer", "options", "output",
		"player", "playlist", "stored_paylist", "update",
	}
}

func closeMpdConn(session *Session, params map[string]string) ([]string, error) {
	logger.Notice("CLOSE")
	session.Close()
	return nil, nil
}

func ping(session *Session, params map[string]string) ([]string, error) {
	logger.Notice("PING")
	return nil, nil
}

func password(session *Session, params map[string]string) ([]string, error) {
	logger.Notice("PASSWORD %s", params["password"])

	if session.Config.Mpd.Password == params["password"] {
		session.Authenticated = true
	} else {
		return nil, MpdAckError{
			Code:    ACK_ERROR_PASSWORD,
			Command: "password",
			Message: "incorrect password",
		}
	}

	return nil, nil
}

func mpdStatus(session *Session, params map[string]string) (result []string, err error) {
	logger.Notice("STATUS")

	result = []string{
		"volume: 100",
		"repeat: 0",
		"random: 0",
		"single: 0",
		"consume: 0",
		"playlist: 1",
		"playlistlength: 0",
		"xfade: 0",
		"state: stop", // pause, play
	}

	return
}

func setIdle(session *Session, params map[string]string) (result []string, err error) {
	logger.Notice("IDLE %s", params)

	// TODO handle subsystems parameter
	subsystems := SUBSYSTEMS
	for _, sub := range subsystems {
		session.addSubscription(sub)
	}

	active := session.getActiveEvents()
	if len(active) == 0 {
		session.preventTimeout = true
		return
	}

	session.clearEvents()
	session.clearSubscriptions()

	for _, subsystem := range active {
		result = append(result, fmt.Sprintf("changed: %s", subsystem))
	}

	return
}

func setNoIdle(session *Session, params map[string]string) (result []string, err error) {
	logger.Notice("NOIDLE")
	session.clearEvents()
	session.clearSubscriptions()
	session.preventTimeout = false
	return
}

func getPlaylistInfo(session *Session, params map[string]string) (result []string, err error) {
	logger.Notice("PLAYLISTINFO")
	return
}
