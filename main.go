package main

import (
	"database/sql"
	"errors"
	"fmt"

	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type OCRResponse struct {
	OCRExitCode                  int            `json:"OCRExitCode"`
	IsErroredOnProcessing        bool           `json:"IsErroredOnProcessing"`
	ErrorMessage                 string         `json:"ErrorMessage"`
	ErrorDetails                 string         `json:"ErrorDetails"`
	SearchablePDFURL             string         `json:"SearchablePDFURL"`
	ProcessingTimeInMilliseconds string         `json:"ProcessingTimeInMilliseconds"`
	ParsedResults                []ParsedResult `json:"ParsedResults"`
}

type ParsedResult struct {
	FileParseExitCode int          `json:"FileParseExitCode"`
	ParsedText        string       `json:"ParsedText"`
	ErrorMessage      string       `json:"ErrorMessage"`
	ErrorDetails      string       `json:"ErrorDetails"`
	TextOverlay       *TextOverlay `json:"TextOverlay,omitempty"`
}

type TextOverlay struct {
	HasOverlay bool   `json:"HasOverlay"`
	Message    string `json:"Message"`
	Lines      []Line `json:"Lines"`
}

type Line struct {
	Words     []Word `json:"Words"`
	MaxHeight int    `json:"MaxHeight"`
	MinTop    int    `json:"MinTop"`
}

type Word struct {
	WordText string `json:"WordText"`
	Left     int    `json:"Left"`
	Top      int    `json:"Top"`
	Height   int    `json:"Height"`
	Width    int    `json:"Width"`
}

type AdvertalystAi struct {
	Summarizer
	TextExtractor
	Db
}

func main() {
	ai := AdvertalystAi{
		TextExtractor: &OcrSpace{
			url:    "https://api.ocr.space/parse/image",
			Client: http.Client{},
		},
		Summarizer: &FaltuSummarizer{},
		Db: &Sqlite{},
	}

	router := mux.NewRouter()

	db, err := sql.Open("sqlite3", "ocr.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	createTables(db)

	router.HandleFunc("/api/v1/page", ai.uploadImage).Methods(http.MethodPost)

	log.Println("Server listening on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func createTables(db *sql.DB) {
	userTable := `
		CREATE TABLE IF NOT EXISTS user (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL
		)
	`
	summaryTable := `
		CREATE TABLE IF NOT EXISTS summary (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			image BLOB NOT NULL,
			ocr_parsed_text TEXT NOT NULL,
			summary TEXT,
			FOREIGN KEY (user_id) REFERENCES user(id)
		)
	`

	if _, err := db.Exec(userTable); err != nil {
		log.Println(err)
	}

	if _, err := db.Exec(summaryTable); err != nil {
		log.Println(err)
	}
}

func getUserIdFromRequest(r *http.Request) string {
	return ""
}

func (ai *AdvertalystAi) uploadImage(w http.ResponseWriter, r *http.Request) {
	image, _, err := r.FormFile("filetype")
	if err != nil {
		log.Println("Error getting image file:", err)
		http.Error(w, "Failed to upload image", http.StatusBadRequest)
		return
	}

	defer image.Close()

	userId := getUserIdFromRequest(r)
	imageId, _ := ai.SaveImage(userId, image)

	log.Println(imageId)

	texts, err := ai.ExtractText(image)
	summaries := make([]string, len(texts))

	for i, text := range texts {
		ai.SaveText(userId, imageId, text)
		if summary, e := ai.Summarize(text); e == nil {
			summaries[i] = summary
			ai.SaveSummary(userId, imageId, summary)
		} else {
			err = errors.Join(err, e)
		}
	}

	log.Println("Summaries: ", summaries)
	log.Println("errors:", err)

	if err != nil {
		log.Println("Error performing OCR:", err)
		http.Error(w, "Failed to perform OCR", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "OCR Result: %s", texts)
}