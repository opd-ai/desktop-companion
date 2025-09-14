package main
package main

import (
	"fmt"
	"encoding/json"
	
	minilmDialog "github.com/opd-ai/minilm/dialog"
)

func main() {
	// Test basic structure access
	config := minilmDialog.LLMConfig{}
	fmt.Printf("LLMConfig: %+v\n", config)
	
	// Test what fields are available by trying to marshal to JSON
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling: %v\n", err)
	} else {
		fmt.Printf("LLMConfig JSON structure:\n%s\n", string(configJSON))
	}
	
	// Test DialogContext
	ctx := minilmDialog.DialogContext{}
	ctxJSON, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling DialogContext: %v\n", err)
	} else {
		fmt.Printf("DialogContext JSON structure:\n%s\n", string(ctxJSON))
	}
}