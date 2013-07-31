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

type CommandHandlerFunc func(context *Session, params map[string]string) ([]string, error)

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

func closeMpdConn(context *Session, params map[string]string) ([]string, error) {
	logger.Notice("CLOSE")
	context.Close()
	return nil, nil
}

func ping(context *Session, params map[string]string) ([]string, error) {
	logger.Notice("PING")
	return nil, nil
}

func password(context *Session, params map[string]string) ([]string, error) {
	logger.Notice("PASSWORD %s", params["password"])

	if context.Config.Mpd.Password == params["password"] {
		// context.Dispatcher.Authenticated = true
	} else {
		return nil, MpdAckError{
			Code:    ACK_ERROR_PASSWORD,
			Command: "password",
			Message: "incorrect password",
		}
	}

	return nil, nil
}

func mpdStatus(context *Session, params map[string]string) (result []string, err error) {
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

func setIdle(context *Session, params map[string]string) (result []string, err error) {
	logger.Notice("IDLE %s", params)

	// TODO handle subsystems parameter
	subsystems := SUBSYSTEMS
	for _, sub := range subsystems {
		context.addSubscription(sub)
	}

	active := context.getActiveEvents()
	if len(active) == 0 {
		context.preventTimeout = true
		return
	}

	context.clearEvents()
	context.clearSubscriptions()

	for _, subsystem := range active {
		result = append(result, fmt.Sprintf("changed: %s", subsystem))
	}

	return
}

func setNoIdle(context *Session, params map[string]string) (result []string, err error) {
	logger.Notice("NOIDLE")
	context.clearEvents()
	context.clearSubscriptions()
	context.preventTimeout = false
	return
}

func getPlaylistInfo(context *Session, params map[string]string) (result []string, err error) {
	logger.Notice("PLAYLISTINFO")
	return
}
