package main

import (
	"regexp"
)

type MpdCommand struct {
	AuthRequired bool
	Pattern *regexp.Regexp
}

var MPD_COMMANDS map[string]MpdCommand

func init() {
	logger.Debug("Initialize protocol")	

	MPD_COMMANDS = map[string]MpdCommand {
		"password": MpdCommand{false, regexp.MustCompile("^password \"(?P<password>[^\"]+)\"$")},
		"test": MpdCommand{true, regexp.MustCompile("^test$")},
	}
}
