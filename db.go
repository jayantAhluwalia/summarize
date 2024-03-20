package main

import (
	"database/sql"
	"io"
)

type Db interface {
	SaveImage(userId string, image io.Reader) (id string, err error)
	SaveText(userId string, imageId string, text string) error
	SaveSummary(userId string, imageId string, summary string) error
}

type Sqlite struct {
	*sql.DB
}

func (db *Sqlite) SaveImage(userId string, image io.Reader) (id string, err error) {
	// TODO
	db.Exec("INSERT INTO summary VALUES(?)")
	return "j", nil
}

func (db *Sqlite) SaveText(userId string, imageId string, text string) error {
	return nil
}

func (db *Sqlite) SaveSummary(userId string, imageId string, summary string) error {
	return nil
}