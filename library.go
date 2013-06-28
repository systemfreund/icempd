package main 

import (
	"os"
	"path/filepath"
	"bitbucket.org/taruti/taglib.go"
)

type Tune struct {
	Uri string
	Title, Artist, Album, Comment, Genre string
	Year, Track int
}

type Library struct {
	path string
	TuneChannel chan Tune
}

func NewLibrary(basePath string) Library {
	result := Library{filepath.Clean(basePath), make(chan Tune, 20)}
	go result.scan()
	return result
}

func (l *Library) scan() {
	logger.Notice("Library path: %s", l.path)
	filepath.Walk(l.path, l.walkFunc)
}

func (l *Library) walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		logger.Error("Error while scanning %s. %s", path, err)
		return err
	}

	if info.Mode().IsRegular() && filepath.Ext(path) == ".mp3" {
		logger.Debug("* %s", path)	
		tune := Tune{
			Uri: path,
		}

		PopulateTune(&tune)

		l.TuneChannel <- tune
	}

	return nil
}

func PopulateTune(tune *Tune) {
	f := taglib.Open(tune.Uri)
	if f == nil {
		// TODO return error
		return
	}
	tags := f.GetTags()
	defer f.Close()

	tune.Title = tags.Title
	tune.Artist = tags.Artist
	tune.Album = tags.Album
	tune.Comment = tags.Comment
	tune.Genre = tags.Genre
	tune.Year = tags.Year
	tune.Track = tags.Track
}