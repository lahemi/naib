package main

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DBFields struct {
	url, title string
	timestamp  int64
	category   string
}

func setupDB() {
	if _, err := os.Stat(dbFile); err != nil {
		db, err := sql.Open("sqlite3", dbFile)
		if err != nil {
			die("Failed to create db file: " + dbFile)
		}
		defer db.Close()

		if _, err = db.Exec(`
            CREATE TABLE IF NOT EXISTS links (
                id INTEGER NOT NULL PRIMARY KEY,
                url TEXT NOT NULL,
                title TEXT,
                timestamp INTEGER NOT NULL, -- UNIX timestamp
                category TEXT,
                FOREIGN KEY(category) REFERENCES categories(category)
            );
            CREATE TABLE IF NOT EXISTS categories (
                category TEXT PRIMARY KEY
            );
            INSERT INTO categories VALUES('music');
            INSERT INTO categories VALUES('img');
            INSERT INTO categories VALUES('lulz');
            INSERT INTO categories VALUES('info');
            INSERT INTO categories VALUES('blank'); -- for no category
        `); err != nil {
			die("Failed to execute SQLite3.")
		}
	}
}

func saveLinksToDB(fields DBFields) {
	// Pretty funky, pattern matching-like.
	switch "" {
	case fields.url:
		stderr("No url to save to DB.")
		return
	case fields.title:
		fields.title = "blank"
	case fields.category:
		fields.category = "blank"
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		stderr("Failed to open the DB.")
		return
	}
	defer db.Close()

	fields.timestamp = int64(time.Now().Unix())

	stmt, err := db.Prepare(`
        INSERT INTO links(url, title, timestamp, category)
        VALUES(?, ?, ?, ?)
    `)
	if err != nil {
		stderr("Failed to prepare SQL statement with error:", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(fields.url, fields.title, fields.timestamp, fields.category)
	if err != nil {
		stderr("Failed to write to DB.")
		return
	}
	stdout(fields.url + " saved to DB.")
}
