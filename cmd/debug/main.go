// debug_test_active_topics.go - Debug program to analyze the TestGetActiveTopics failure
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/opd-ai/desktop-companion/lib/dialog"
)

func main() {
	ctx := dialog.NewConversationContext()

	// Add message to create topic
	err := ctx.AddMessage(context.Background(), "The weather is sunny")
	if err != nil {
		fmt.Printf("AddMessage failed: %v\n", err)
		return
	}

	fmt.Printf("After AddMessage:\n")
	fmt.Printf("Total topics: %d\n", len(ctx.Topics))
	for i, topic := range ctx.Topics {
		fmt.Printf("Topic %d: %s, confidence: %.2f, timestamp: %v\n",
			i, topic.Name, topic.Confidence, topic.LastSeen)
	}

	// Should have active topics
	active := ctx.GetActiveTopics()
	fmt.Printf("Active topics before timestamp change: %d\n", len(active))

	// Make topic old by manipulating timestamp
	if len(ctx.Topics) > 0 {
		fmt.Printf("Changing topic 0 timestamp from %v to %v\n",
			ctx.Topics[0].LastSeen, time.Now().Add(-10*time.Minute))
		ctx.Topics[0].LastSeen = time.Now().Add(-10 * time.Minute)
	}

	fmt.Printf("After timestamp change:\n")
	for i, topic := range ctx.Topics {
		fmt.Printf("Topic %d: %s, confidence: %.2f, timestamp: %v\n",
			i, topic.Name, topic.Confidence, topic.LastSeen)
	}

	// Should have no active topics now
	active = ctx.GetActiveTopics()
	fmt.Printf("Active topics after timestamp change: %d\n", len(active))

	cutoff := time.Now().Add(-5 * time.Minute)
	fmt.Printf("Cutoff time: %v\n", cutoff)

	for i, topic := range ctx.Topics {
		after := topic.LastSeen.After(cutoff)
		fmt.Printf("Topic %d (%s) LastSeen: %v, After cutoff: %t\n",
			i, topic.Name, topic.LastSeen, after)
	}
}
