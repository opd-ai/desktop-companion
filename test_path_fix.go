package main

import (
	"fmt"
	"github.com/opd-ai/desktop-companion/internal/character"
)

func main() {
	card, err := character.LoadCard("assets/characters/default/character.json")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: loaded %s\n", card.Name)
	}
}
