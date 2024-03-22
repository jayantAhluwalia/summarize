package main

import (
	// "bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Db interface {
	SaveImage(userId int64, image []byte) (id int64, err error)
	SaveUser(userName string) (id int64, err error)
	SaveText(userId int64, text string) error
	SaveSummary(userId int64, summary string) error
	GetUserId(username string) (id int64, found bool)
	GetSummaryById(imageId string) (imagePath string, ocrText string, summary string, err error)
	GetAllIds(userId string) (ids []string, err error)
}

type Sqlite struct {
	*sql.DB
	imageDirPath string
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

func (db *Sqlite) SaveImage(userId int64, image []byte) (id int64, err error) {
	now := time.Now()
	timestamp := now.Format("2006-01-02_15-04-05")
	fileName := filepath.Join(db.imageDirPath, fmt.Sprintf("%s_image.jpg", timestamp))

	file, err := os.Create(fileName)
	if err != nil {
		return id, err
	}
	defer file.Close()

	_, err = file.Write(image)
	if err != nil {
		return id, err
	}

  imageURL := fmt.Sprintf("http://localhost:8000/%s", fileName)


	stmt, err := db.Prepare("INSERT INTO summary (user_id, image_path) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(userId, imageURL)
	if err != nil {
		return 0, err
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

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

func (db *Sqlite) SaveSummary(userId int64, summary string) error {
	stmt, err := db.Prepare("UPDATE summary SET summary = ? WHERE user_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(summary, userId)

	if err != nil {
		return err
	}
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

func (db *Sqlite) GetSummaryById(imageId string) (imagePath string, ocrText string, summary string, err error) {
	// Build the SQL query
	query := `
    SELECT image_path, ocr_parsed_text, summary
    FROM summary
    WHERE id = ?;
  `

	// Prepare the statement
	stmt, err := db.Prepare(query)
	if err != nil {
		return "", "", "", fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close() // Close the statement after use

	// Execute the statement with the userId parameter
	rows, err := stmt.Query(imageId)
	if err != nil {
		return "", "", "", fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close() // Close the rows after use

	// Check if a row was found
	if !rows.Next() {
		return "", "", "", fmt.Errorf("no record found for user ID: %s", imageId)
	}

	// Scan the retrieved data
	err = rows.Scan(&imagePath, &ocrText, &summary)
	if err != nil {
		return "", "", "", fmt.Errorf("error scanning data: %w", err)
	}

	return imagePath, ocrText, summary, nil
}

func (db *Sqlite) GetAllIds(userId string) (ids []string, err error) {
	query := `
	SELECT id
	FROM summary
	WHERE user_id = ?;
`

	// Prepare the statement
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close() // Close the statement after use

	// Execute the statement with the userId parameter
	rows, err := stmt.Query(userId)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close() // Close the rows after use

	// Initialize an empty slice for IDs
	ids = []string{}

	// Scan each row and append the ID to the slice
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("error scanning data: %w", err)
		}
		ids = append(ids, id)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ids, nil
}
