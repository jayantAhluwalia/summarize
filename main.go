package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	"log"
	"net/http"
	"github.com/gorilla/mux"
)

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

	return string(body), nil
}