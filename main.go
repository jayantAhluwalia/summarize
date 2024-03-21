package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"

	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sashabaranov/go-openai"
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

func buildSummarizer() *GptSummarizer {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	openAiAuthToken := os.Getenv("OPENAI_API_KEY")
	openAiClient := openai.NewClient(openAiAuthToken)

	return &GptSummarizer{openAiClient}
}

const imageDirPath string = "uploads"

func setupDb() *sql.DB {
	dbFile := "ocr.db"

	// create db file if not exists
	f, err := os.OpenFile(dbFile, os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	f.Close()

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	os.Mkdir("uploads", 0755)

	return db
}

func main() {
	db := setupDb()

	defer db.Close()

	ai := AdvertalystAi{
		TextExtractor: &OcrSpace{
			url:    "https://api.ocr.space/parse/image",
			Client: http.Client{},
		},
		Summarizer: &FaltuSummarizer{},
		Db:         &Sqlite{db, imageDirPath},
	}

	router := mux.NewRouter()

	createTables(db)

	router.HandleFunc("/api/v1/page", ai.uploadImage).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/page/{id}", ai.getPageById).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/page/{id}/all", ai.getAllIds).Methods(http.MethodGet)

	log.Println("Server listening on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func createTables(db *sql.DB) {
	userTable := `
		CREATE TABLE IF NOT EXISTS user (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE
		)
	`
	summaryTable := `
		CREATE TABLE IF NOT EXISTS summary (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			image_path TEXT NOT NULL,
			ocr_parsed_text TEXT,
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
	return r.FormValue("userName")
}

func (ai *AdvertalystAi) uploadImage(w http.ResponseWriter, r *http.Request) {
	log.Println("hellhe")
	response := struct {
		Success bool   `json:"success"`
		UserId  string `json:"user_id"`
	}{}

	image, _, err := r.FormFile("filetype")
	if err != nil {
		log.Println("Error getting image file:", err)
		http.Error(w, "Failed to upload image", http.StatusBadRequest)
		return
	}

	defer image.Close()
	userName := getUserIdFromRequest(r)

	userId, found := ai.GetUserId(userName)
	if !found {
		userId, _ = ai.SaveUser(userName)
	}

	response.UserId = strconv.FormatInt(userId, 10)
	imageBytes, _ := io.ReadAll(image)

	log.Println("user id:", userId)

	imageId, err := ai.SaveImage(userId, imageBytes)
	if err != nil {
		log.Println("error saving image: ", err)
		return
	}
	log.Println(imageId)

	texts, err := ai.ExtractText(imageBytes)
	summaries := make([]string, len(texts))

	for i, text := range texts {
		if err := ai.SaveText(userId, text); err != nil {
			log.Println("error saving ocr text: ", err)
			return
		}

		if summary, e := ai.Summarize(text); e == nil {
			summaries[i] = summary
			ai.SaveSummary(userId, summary)
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

	response.Success = true
	responseBytes, _ := json.Marshal(response)
	_, _ = w.Write(responseBytes)
}

func (ai *AdvertalystAi) getPageById(w http.ResponseWriter, r *http.Request) {
	response := struct {
		ImageURL      string `json:"imageURL"`
		TextExtracted string `json:"textExtracted"`
		TextSummary   string `json:"textSummary"`
	}{}

	idStr := mux.Vars(r)["id"]

	imagePath, ocrText, summary, err := ai.GetSummaryById(idStr)

	if err != nil {
		log.Println("Error in getting summary:", err)
		http.Error(w, "Failed to get summary", http.StatusInternalServerError)
		return
	}

	response.ImageURL = imagePath
	response.TextExtracted = ocrText
	response.TextSummary = summary

	responseBytes, _ := json.Marshal(response)
	_, _ = w.Write(responseBytes)
}

func (ai *AdvertalystAi) getAllIds(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Ids []string `json:"ids"`
	}{}

	idStr := mux.Vars(r)["id"]

	allIds, err := ai.GetAllIds(idStr)

	if err != nil {
		log.Println("Error in getting all Ids:", err)
		http.Error(w, "Failed to get all Ids", http.StatusInternalServerError)
		return
	}

	response.Ids = allIds

	responseBytes, _ := json.Marshal(response)
	_, _ = w.Write(responseBytes)
}
