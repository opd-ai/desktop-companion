package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	
	"github.com/opd-ai/minilm/dialog"
)

func main() {
	// Examine the actual structure of LLMConfig
	config := dialog.LLMConfig{}
	t := reflect.TypeOf(config)
	
	fmt.Printf("LLMConfig fields:\n")
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fmt.Printf("  %s: %s\n", field.Name, field.Type)
	}
	
	// Also examine DialogContext
	ctx := dialog.DialogContext{}
	t2 := reflect.TypeOf(ctx)
	
	fmt.Printf("\nDialogContext fields:\n")
	for i := 0; i < t2.NumField(); i++ {
		field := t2.Field(i)
		fmt.Printf("  %s: %s\n", field.Name, field.Type)
	}
}