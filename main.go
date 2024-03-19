package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type OCRResponse struct {
  OCRExitCode                int                 `json:"OCRExitCode"`
  IsErroredOnProcessing     bool                   `json:"IsErroredOnProcessing"`
  ErrorMessage              string                 `json:"ErrorMessage"`
  ErrorDetails              string                 `json:"ErrorDetails"`
  SearchablePDFURL          string                 `json:"SearchablePDFURL"`
  ProcessingTimeInMilliseconds string                 `json:"ProcessingTimeInMilliseconds"`
  ParsedResults             []ParsedResult         `json:"ParsedResults"`
}

type ParsedResult struct {
  FileParseExitCode        int                 `json:"FileParseExitCode"`
  ParsedText               string                 `json:"ParsedText"`
  ErrorMessage              string                 `json:"ErrorMessage"`
  ErrorDetails              string                 `json:"ErrorDetails"`
  TextOverlay               *TextOverlay           `json:"TextOverlay,omitempty"`
}

type TextOverlay struct {
  HasOverlay               bool                   `json:"HasOverlay"`
  Message                  string                 `json:"Message"`
  Lines                    []Line                 `json:"Lines"`
}

type Line struct {
  Words                     []Word                 `json:"Words"`
  MaxHeight                 int                    `json:"MaxHeight"`
  MinTop                   int                    `json:"MinTop"`
}

type Word struct {
  WordText                 string                 `json:"WordText"`
  Left                      int                    `json:"Left"`
  Top                       int                    `json:"Top"`
  Height                    int                    `json:"Height"`
  Width                     int                    `json:"Width"`
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/page", uploadImage).Methods("POST")

	log.Println("Server listening on port 8000") 
	log.Fatal(http.ListenAndServe(":8000", router))
}


func uploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("Received non-POST request for /api/v1/page")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("filetype")
	if err != nil {
		log.Println("Error getting image file:", err)
		http.Error(w, "Failed to upload image", http.StatusBadRequest)
		return
	}

	text, err := performOCR(file)

	defer file.Close()

	if err != nil {
		log.Println("Error performing OCR:", err)
		http.Error(w, "Failed to perform OCR", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "OCR Result: %s", text)
}


func performOCR(file multipart.File) (string, error) {
	url := "https://api.ocr.space/parse/image"
	method := "POST"

	writer := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(writer)

	if err := multipartWriter.WriteField("language", "eng"); err != nil {
			return "", err
	}

	if err := multipartWriter.WriteField("isOverlayRequired", "false"); err != nil {
		return "", err
	}
	if err := multipartWriter.WriteField("iscreatesearchablepdf", "false"); err != nil {
		return "", err
	}
	if err := multipartWriter.WriteField("issearchablepdfhidetextlayer", "false"); err != nil {
		return "", err
	}

	part, err := multipartWriter.CreateFormFile("filetype", "/pnp/image.png")
	if err != nil {
			return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
			return "", err
	}
	_ = multipartWriter.Close()

	contentType := multipartWriter.FormDataContentType()

	req, err := http.NewRequest(method, url, writer)
	if err != nil {
			return "", err
	}
	req.Header.Set("apikey", "helloworld")
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
			return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
			return "", err
	}

	var response OCRResponse

	err = json.Unmarshal(body, &response)

	if err != nil {
    return "", errors.New("Failed to parse OCR response")
  }

	if response.IsErroredOnProcessing == true {
		return response.ErrorMessage, nil
	}

	results := response.ParsedResults

	return string(results[0].ParsedText), nil
}