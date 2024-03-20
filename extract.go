package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

type TextExtractor interface {
	ExtractText(image io.Reader) ([]string, error)
}

type OcrSpace struct {
	url string
	http.Client
}

func (space *OcrSpace) ExtractText(image io.Reader) (texts []string, err error) {
	payload := new(bytes.Buffer)
	writer := multipart.NewWriter(payload)

	config := map[string]string{
		"language":                     "eng",
		"isOverlayRequired":            "false",
		"iscreatesearchablepdf":        "false",
		"issearchablepdfhidetextlayer": "false",
	}

	for key, value := range config {
		if e := writer.WriteField(key, value); e != nil {
			err = errors.Join(err, e)
		}
	}

	if err != nil {
		return texts, err
	}

	part, err := writer.CreateFormFile("filetype", "image.png")
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, image); err != nil {
		return nil, err
	}

	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, space.url, payload)
	if err != nil {
		return texts, err
	}

	req.Header.Set("apikey", "K85721292588957")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := space.Client.Do(req)
	if err != nil {
		return texts, err
	}

	defer resp.Body.Close()

	var response OCRResponse

	// if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
	// 	return texts, errors.New("failed to parse OCR response")
	// }

	body, err := io.ReadAll(resp.Body)
	if err != nil {
			return nil, err
	}


	if err := json.Unmarshal(body, &response); err != nil {
		log.Println(string(body))
		return nil, errors.New("failed to parse OCR response")
	}

	if response.IsErroredOnProcessing {
		return texts, errors.New("ocr space server error")
	}

	for _, result := range response.ParsedResults {
		texts = append(texts, result.ParsedText)
	}

	return texts, nil
}
