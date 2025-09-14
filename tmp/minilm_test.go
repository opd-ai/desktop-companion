package main

import (
	"fmt"
	"github.com/opd-ai/minilm/dialog"
)

func main() {
	// Explore the miniLM dialog package API
	fmt.Println("miniLM dialog package imported successfully")
	
	// Try to list available types
	fmt.Printf("DialogManager type: %T\n", &dialog.DialogManager{})
	fmt.Printf("LLMBackend type: %T\n", &dialog.LLMBackend{})
}