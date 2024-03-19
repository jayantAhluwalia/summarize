package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	// "io/ioutil"
	"log"
	"net/http"

	// "github.com/disintegration/imaging"
	// "github.com/google/uuid"
	"github.com/gorilla/mux"
	// "github.com/otiai10/gosseract/v2"
)

func main() {
	router := mux.NewRouter()

	// Register the route with the correct HTTP method (POST)
	router.HandleFunc("/api/v1/page", uploadImage).Methods("POST")

	log.Println("Server listening on port 8000") // Informational message on startup
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

	// filename := fmt.Sprintf("%s.png", uuid.New().String()) // Example using uuid

  // Create a destination path
  // destinationPath := fmt.Sprintf("./uploads/%s", filename)

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
	defer multipartWriter.Close()

	// Write additional form fields if needed (e.g., language)
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

	// Write image data directly
	part, err := multipartWriter.CreateFormFile("filetype", "/pnp/image.png")
	if err != nil {
			return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
			return "", err
	}

	contentType := multipartWriter.FormDataContentType()

	log.Print(contentType)
	// Create HTTP request
	req, err := http.NewRequest(method, url, writer)
	if err != nil {
			return "", err
	}
	req.Header.Set("apikey", "helloworld") // Replace with your actual API key
	req.Header.Set("Content-Type", contentType)

	// Execute request and process response
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

	// Parse response based on OCR.space API format (handle potential errors)
	return string(body), nil
}