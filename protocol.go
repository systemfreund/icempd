package main

import (
	"regexp"
)

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

type MpdCommand struct {
	AuthRequired bool
	Pattern *regexp.Regexp
}

var MPD_COMMANDS map[string]MpdCommand

func init() {
	logger.Debug("Initialize protocol")	

	MPD_COMMANDS = map[string]MpdCommand {
		"ping": MpdCommand{false, regexp.MustCompile("^ping$")},
		"password": MpdCommand{false, regexp.MustCompile("^password \"(?P<password>[^\"]+)\"$")},
		"test": MpdCommand{true, regexp.MustCompile("^test$")},
	}
}

func ping() {
	
}