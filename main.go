package main

import (
	"context"
	_ "embed"
	"log"
	"os"
	"os/exec"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	apiBaseURL = "https://api.deepseek.com" // e.g. "https://api.openai.com/v1"
	apiKeyEnv  = "DEEPSEEK_TOKEN" // e.g. "OPENAI_API_KEY"
	model      = "deepseek-v4-flash" // e.g. "gpt-4o-mini"
)

//go:embed prompt.md
var instructions []byte

//go:embed prompt_description.md
var instructions_description []byte

var instructionsContent []byte

func main() {
	var usingDescription = false

	if len(os.Args) > 1 {
		if os.Args[1] == "-d" {
			usingDescription = true
		}
	}

	fileTmp, err := os.CreateTemp("", "aicommit-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(fileTmp.Name())

	process := exec.Command("git", "diff", "--staged")

	output, err := process.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	diffContent := string(output)
	if diffContent == "" {
		log.Fatal("No staged changes")
	}

	if apiKeyEnv == "" {
		log.Fatal("apiKeyEnv is not configured")
	}

	apiKey := os.Getenv(apiKeyEnv)
	if apiKey == "" {
		log.Fatalf("No API key set (env: %s)", apiKeyEnv)
	}

	if apiBaseURL == "" {
		log.Fatal("apiBaseURL is not configured")
	}

	if model == "" {
		log.Fatal("model is not configured")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = apiBaseURL

	client := openai.NewClientWithConfig(config)

	if usingDescription {
		instructionsContent = instructions_description
	} else {
		instructionsContent = instructions
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: string(instructionsContent),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: diffContent,
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	if len(resp.Choices) == 0 {
		log.Fatal("Empty response from API")
	}

	commitMessage := strings.TrimSpace(resp.Choices[0].Message.Content)
	if commitMessage == "" {
		log.Fatal("Empty commit message from API")
	}

	fileTmp.Write([]byte(commitMessage))

	cmd := exec.Command("git", "commit", "-F", fileTmp.Name(), "--edit")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
