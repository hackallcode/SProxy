package repeater

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

func getRequest(id int64) (req request, err error) {
	row := db.QueryRow("SELECT uri, method, header, body FROM requests WHERE id = $1", id)
	err = row.Scan(&req.uri, &req.method, &req.header, &req.body)
	return req, err
}
