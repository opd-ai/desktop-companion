package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// DataProcessor handles concurrent data processing
type DataProcessor struct {
	mu      sync.Mutex
	results []string
}

// ProcessConcurrently processes data items concurrently
func (dp *DataProcessor) ProcessConcurrently(items []string) []string {
	// BUG: Missing context timeout handling and unbuffered channel causes deadlock
	ch := make(chan string) // Unbuffered channel - potential deadlock

	for _, item := range items {
		go func(data string) {
			// Simulate processing time - longer delay to increase deadlock probability
			time.Sleep(50 * time.Millisecond)
			processed := "processed_" + data
			ch <- processed // May block indefinitely if no reader or if reader is blocked
		}(item)
	}

	// BUG: No timeout handling, will hang if goroutines don't complete or get blocked
	var results []string
	for i := 0; i < len(items); i++ {
		select {
		case result := <-ch:
			results = append(results, result)
		case <-time.After(100 * time.Millisecond): // This creates a race condition
			// BUG: This timeout is too short and may cause partial results
			break
		}
	}

	return results
}

// ProcessConcurrentlyWithContext processes data with proper context handling
func (dp *DataProcessor) ProcessConcurrentlyWithContext(ctx context.Context, items []string) ([]string, error) {
	// Fixed: Buffered channel and context handling
	ch := make(chan string, len(items))
	defer close(ch)

	var wg sync.WaitGroup
	wg.Add(len(items))

	for _, item := range items {
		go func(data string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				// Simulate processing time
				time.Sleep(10 * time.Millisecond)
				processed := "processed_" + data
				select {
				case ch <- processed:
				case <-ctx.Done():
					return
				}
			}
		}(item)
	}

	// Wait for all goroutines to complete or context timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	var results []string
	for {
		select {
		case result, ok := <-ch:
			if ok {
				results = append(results, result)
			}
		case <-done:
			return results, nil
		case <-ctx.Done():
			return results, ctx.Err()
		}

		if len(results) == len(items) {
			return results, nil
		}
	}
}

// TestDataProcessorConcurrent tests concurrent processing without timeout handling
func TestDataProcessorConcurrent(t *testing.T) {
	processor := &DataProcessor{}

	// Test with larger dataset that is more likely to cause the deadlock
	items := make([]string, 10)
	for i := 0; i < 10; i++ {
		items[i] = fmt.Sprintf("item%d", i+1)
	}

	// BUG: This test will hang due to concurrency issues in the implementation
	results := processor.ProcessConcurrently(items)

	if len(results) != len(items) {
		t.Errorf("Expected %d results, got %d", len(items), len(results))
	}

	// Verify all items were processed
	for _, result := range results {
		if result[:10] != "processed_" {
			t.Errorf("Expected result to start with 'processed_', got %s", result)
		}
	}
}

// TestDataProcessorWithContext tests concurrent processing with proper timeout handling
func TestDataProcessorWithContext(t *testing.T) {
	processor := &DataProcessor{}

	// Test with proper context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	items := []string{"item1", "item2", "item3"}

	results, err := processor.ProcessConcurrentlyWithContext(ctx, items)
	if err != nil {
		t.Fatalf("ProcessConcurrentlyWithContext failed: %v", err)
	}

	if len(results) != len(items) {
		t.Errorf("Expected %d results, got %d", len(items), len(results))
	}

	// Verify all items were processed
	for _, result := range results {
		if result[:10] != "processed_" {
			t.Errorf("Expected result to start with 'processed_', got %s", result)
		}
	}
}
