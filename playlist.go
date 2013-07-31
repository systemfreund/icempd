package main

import (
	"github.com/op/go-logging"
)

type PlaylistEntry struct {
	id   int
	tune Tune
}

type Playlist struct {
	version int
	entries []PlaylistEntry

	*logging.Logger
}

func (p *Playlist) add(entry PlaylistEntry) {
	p.entries = append(p.entries, entry)

	p.increaseVersion()
}

func (p *Playlist) increaseVersion() {
	p.version += 1
	// TODO trigger tracklist changed
}
