package sproxy

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

type request struct {
	id     int64
	uri    string
	method string
	header []byte
	body   []byte
}

func initDb() {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalln(err)
	}
}

func insertRequest(req request) (err error) {
	_, err = db.Exec("INSERT INTO requests (uri, method, header, body) VALUES ($1, $2, $3, $4)",
		req.uri, req.method, req.header, req.body)
	return
}
