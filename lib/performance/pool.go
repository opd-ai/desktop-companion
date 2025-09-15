// pool.go: Memory pools for performance optimization
// Uses sync.Pool to reduce garbage collection pressure for frequently allocated objects
// Focuses on character state, animation frames, and network message types

// Package performance provides memory optimization utilities for the desktop companion application.
// It includes object pools using sync.Pool to reduce garbage collection pressure for frequently
// allocated types like character state, animation frames, and network messages.
package performance

import (
	"sync"
)

// CharacterState represents character state data that gets frequently allocated
type CharacterState struct {
	ID       string
	Health   int
	Energy   int
	Mood     string
	Activity string
}

// AnimationFrame represents a single animation frame with timing
type AnimationFrame struct {
	ImageData []byte
	Duration  int64
	Width     int
	Height    int
}

// NetworkMessage represents network protocol messages
type NetworkMessage struct {
	Type    string
	Payload []byte
	Sender  string
	Target  string
}

// Global pools for commonly allocated objects
var (
	// characterStatePool reduces allocations for character state updates
	characterStatePool = sync.Pool{
		New: func() interface{} {
			return &CharacterState{}
		},
	}

	// animationFramePool reduces allocations during animation processing
	animationFramePool = sync.Pool{
		New: func() interface{} {
			return &AnimationFrame{
				ImageData: make([]byte, 0, 4096), // pre-allocate common size
			}
		},
	}

	// networkMessagePool reduces allocations for network communication
	networkMessagePool = sync.Pool{
		New: func() interface{} {
			return &NetworkMessage{
				Payload: make([]byte, 0, 1024), // pre-allocate common size
			}
		},
	}
)

// GetCharacterState returns a reused CharacterState from the pool.
// Must be returned using PutCharacterState when done to prevent memory leaks.
func GetCharacterState() *CharacterState {
	return characterStatePool.Get().(*CharacterState)
}

// PutCharacterState returns a CharacterState to the pool after clearing sensitive data.
// This prevents memory leaks and allows the object to be safely reused.
func PutCharacterState(cs *CharacterState) {
	// Clear data to prevent memory leaks
	cs.ID = ""
	cs.Health = 0
	cs.Energy = 0
	cs.Mood = ""
	cs.Activity = ""
	characterStatePool.Put(cs)
}

// GetAnimationFrame returns a reused AnimationFrame from the pool.
// Must be returned using PutAnimationFrame when done to prevent memory leaks.
func GetAnimationFrame() *AnimationFrame {
	return animationFramePool.Get().(*AnimationFrame)
}

// PutAnimationFrame returns an AnimationFrame to the pool after clearing data.
// This prevents memory leaks and resets the slice capacity for reuse.
func PutAnimationFrame(af *AnimationFrame) {
	// Clear data to prevent memory leaks
	af.ImageData = af.ImageData[:0] // reset slice but keep capacity
	af.Duration = 0
	af.Width = 0
	af.Height = 0
	animationFramePool.Put(af)
}

// GetNetworkMessage returns a reused NetworkMessage from the pool.
// Must be returned using PutNetworkMessage when done to prevent memory leaks.
func GetNetworkMessage() *NetworkMessage {
	return networkMessagePool.Get().(*NetworkMessage)
}

// PutNetworkMessage returns a NetworkMessage to the pool after clearing data.
// This prevents memory leaks and resets the slice capacity for reuse.
func PutNetworkMessage(nm *NetworkMessage) {
	// Clear data to prevent memory leaks
	nm.Type = ""
	nm.Payload = nm.Payload[:0] // reset slice but keep capacity
	nm.Sender = ""
	nm.Target = ""
	networkMessagePool.Put(nm)
}