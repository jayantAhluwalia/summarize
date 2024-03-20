package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	// "github.com/sashabaranov/go-openai"
	openai "github.com/sashabaranov/go-openai"
)

type Summarizer interface {
	Summarize(string, bool) (string, error)
}

type FaltuSummarizer struct{}

func (*FaltuSummarizer) Summarize(text string, useGpt bool) (string, error) {
	if useGpt {
		return "this is a random summary, to avoid running out of source", nil
	}
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	c := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()
	content := fmt.Sprintf("Please summarize the provided text in less than 30 characters: %s", text)
	
	dialogue := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: content},
	}
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		MaxTokens: 1000,
		Messages: dialogue,
	}

	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Println("Completion error: ", err)
		return "", err
	}

	summary := resp.Choices[0]
	log.Println("sum mess: ", summary.Message)
	return summary.Message.Content, nil
}
