//go:build ignore

package main

import (
	"flag"
	"log"

	"github.com/opd-ai/desktop-companion/internal/embedding"
)

var (
	characterName = flag.String("character", "", "Character name to embed")
	outputDir     = flag.String("output", "", "Output directory for generated code")
)

func main() {
	flag.Parse()

	if *characterName == "" || *outputDir == "" {
		log.Fatal("Both -character and -output flags are required")
	}

	// Generate embedded character application
	if err := embedding.GenerateEmbeddedCharacter(*characterName, *outputDir); err != nil {
		log.Fatalf("Failed to generate embedded application: %v", err)
	}
}
