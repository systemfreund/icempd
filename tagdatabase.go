package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteTagDb struct {
	Path string
	TuneChannel <-chan Tune

	db *sql.DB
}

func (db *SqliteTagDb) Close() {
	db.Close()
}

func (db *SqliteTagDb) populate() {
	for {
		tune := <- db.TuneChannel

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
	logger.Notice("Sqlite tag database at %s", path)
	
	result := *new(SqliteTagDb)
	db, err := sql.Open("sqlite3", path)
	if err != nil {	panic("Can't open tag database") }
	result.db = db
	result.TuneChannel = tuneChannel

	go result.populate()
	return result
}

