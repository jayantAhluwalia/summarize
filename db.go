package main

import (
	"database/sql"
	"io"
	"log"
)

type Db interface {
	SaveImage(userId int64, image io.Reader) (id int64, err error)
	SaveUser(userName string) (id int64, err error)
	SaveText(userId int64, imageId int64, text string) error
	SaveSummary(userId int64, imageId int64, summary string) error
	GetUserId(username string) (id int64, found bool)
}

type Sqlite struct {
	*sql.DB
}

func (db *Sqlite) SaveUser(userName string) (id int64, err error) {
	stmt, err := db.Prepare("INSERT INTO user (username) VALUES (?)")
	if err != nil {
		return id, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(userName)
	if err != nil {
		log.Println("error db exc:", err)
		return id, err
	}

	userId, err := result.LastInsertId()
	if err != nil {
		return id, err
	}

	return userId, nil
}

func (db *Sqlite) SaveImage(userId int64, image io.Reader) (id int64, err error) {
	// TODO
	db.Exec("INSERT INTO summary (image) VALUES(?)")
	// stmt, err := db.Prepare("insert into image")
	return id, nil
}

func (db *Sqlite) SaveText(userId int64, imageId int64, text string) error {
	return nil
}

func (db *Sqlite) SaveSummary(userId int64, imageId int64, summary string) error {
	return nil
}

func (db *Sqlite) GetUserId(username string) (id int64, found bool) {
	stmt, err := db.Prepare("SELECT id FROM user WHERE username = ?")
	if err != nil {
		return id, false
	}
	defer stmt.Close()

	var userId int64
	err = stmt.QueryRow(username).Scan(&userId)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false
		}
		return 0, false
	}

	return userId, true
}