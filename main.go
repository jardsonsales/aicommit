package main

import (
	"context"
	_ "embed"
	"log"
	"os"
	"os/exec"

	"google.golang.org/genai"
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
		return
	}

	diffContent := string(output)
	if(diffContent == "") {
		log.Fatal("No staged changes")
		return
	}

	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Fatal("No Gemini API KEY set")
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// diff, err := os.ReadFile("diff.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }

	if usingDescription == false {
		instructionsContent = instructions
	} else {
		instructionsContent = instructions_description
	}

	config := &genai.GenerateContentConfig{

		SystemInstruction: genai.NewContentFromText(string(instructionsContent), genai.RoleUser),
	}

	result, _ := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash-lite",
		genai.Text(diffContent),
		config,
	)

	fileTmp.Write([]byte(result.Text()))

	cmd := exec.Command("git", "commit", "-F", fileTmp.Name(), "--edit")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
