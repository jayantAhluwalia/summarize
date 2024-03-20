package main

type Summarizer interface {
	Summarize(string) (string, error)
}

type FaltuSummarizer struct{}

func (*FaltuSummarizer) Summarize(text string) (string, error) {
	 return text[:4], nil
}