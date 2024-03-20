package main

import (
	"context"
	"fmt"
	"log"
	openai "github.com/sashabaranov/go-openai"
)

type Summarizer interface {
	Summarize(string) (string, error)
}

type FaltuSummarizer struct{}

func (*FaltuSummarizer) Summarize(text string) (string, error) {
	return text[:4], nil
}

type GptSummarizer struct{
	openAiClient *openai.Client	
}

func (gpt *GptSummarizer) Summarize(text string) (string, error) {
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

	resp, err := gpt.openAiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Println("Completion error: ", err)
		return "", err
	}

	summary := resp.Choices[0]
	log.Println("sum mess: ", summary.Message)
	return summary.Message.Content, nil
}