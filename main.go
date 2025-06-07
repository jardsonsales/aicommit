package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	_ "embed"

	"google.golang.org/genai"
)

//go:embed prompt.md
var instructions []byte

func main() {

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

	ctx := context.Background()
	client, _ := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})

	// diff, err := os.ReadFile("diff.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(string(instructions), genai.RoleUser),
	}

	result, _ := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
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
