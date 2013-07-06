package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/op/go-logging"
)

type SqliteTagDb struct {
	Path string
	TuneChannel <-chan Tune

	db *sql.DB

	*logging.Logger
}

func (db *SqliteTagDb) Close() {
	db.Close()
}

func (db *SqliteTagDb) populate() {
	for tune := range db.TuneChannel {
		tx, err := db.db.Begin()
		if err != nil { panic("Can't create transaction") }

		stmt, err := tx.Prepare("insert into Tunes(Uri, Title, Artist, Album) values (?, ?, ?, ?)")
		if err != nil { panic("Can't prepare statement") }
		defer stmt.Close()

		stmt.Exec(tune.Uri, tune.Title, tune.Artist, tune.Album)

		tx.Commit()
	}
}

func NewSqliteTagDb(path string, tuneChannel <-chan Tune) SqliteTagDb {
	result := *new(SqliteTagDb)
	result.Logger = logging.MustGetLogger(LOGGER_NAME)
	db, err := sql.Open("sqlite3", path)
	if err != nil {	panic("Can't open tag database") }
	result.db = db
	result.TuneChannel = tuneChannel

	result.Notice("Sqlite tag database at %s", path)

	go result.populate()
	return result
}

