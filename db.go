package main

import (
	// "bytes"
	"database/sql"
	"io"
	"log"
)

type Db interface {
	SaveImage(userId int64, image io.Reader) (id int64, err error)
	SaveUser(userName string) (id int64, err error)
	SaveText(userId int64, text string) error
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

func (db *Sqlite) SaveImage(userId int64, image io.Reader) (int64, error) {
  // 1. Prepare the SQL statement with placeholders for user ID and image data
  stmt, err := db.Prepare("INSERT INTO summary (user_id, image) VALUES (?, ?)")
  if err != nil {
    return 0, err
  }
  defer stmt.Close() // Close the prepared statement after use


  // 2. Convert reader to byte slice (assuming image data is small)
  imageData, _ := io.ReadAll(image)
  // 3. Execute the statement with user ID and image data
  result, err := stmt.Exec(userId, imageData)
  if err != nil {
    return 0, err
  }

  // 4. Get the ID of the inserted image row (assuming auto-increment)
  insertedID, err := result.LastInsertId()
  if err != nil {
    return 0, err
  }

  // 5. Return the generated image ID
  return insertedID, nil
}

func (db *Sqlite) SaveText(userId int64, text string) error {
	stmt, err := db.Prepare("UPDATE summary SET ocr_parsed_text = ? WHERE user_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(text, userId)

	if err != nil {
		return err
	}
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
