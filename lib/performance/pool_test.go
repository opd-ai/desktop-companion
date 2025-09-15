// pool_test.go: Benchmark tests for memory pool performance
// Measures allocation reduction and performance improvement

package performance

import (
	"testing"
)

func TestCharacterStatePool(t *testing.T) {
	cs := GetCharacterState()
	if cs == nil {
		t.Fatal("expected non-nil CharacterState")
	}
	cs.ID = "test"
	cs.Health = 100
	PutCharacterState(cs)
	
	// Get another one to verify pool reuse
	cs2 := GetCharacterState()
	if cs2.ID != "" || cs2.Health != 0 {
		t.Error("CharacterState not properly cleared")
	}
	PutCharacterState(cs2)
}

func TestAnimationFramePool(t *testing.T) {
	af := GetAnimationFrame()
	if af == nil {
		t.Fatal("expected non-nil AnimationFrame")
	}
	af.ImageData = append(af.ImageData, 1, 2, 3)
	af.Width = 64
	PutAnimationFrame(af)
	
	af2 := GetAnimationFrame()
	if len(af2.ImageData) != 0 || af2.Width != 0 {
		t.Error("AnimationFrame not properly cleared")
	}
	PutAnimationFrame(af2)
}

func TestNetworkMessagePool(t *testing.T) {
	nm := GetNetworkMessage()
	if nm == nil {
		t.Fatal("expected non-nil NetworkMessage")
	}
	nm.Type = "test"
	nm.Payload = append(nm.Payload, 1, 2, 3)
	PutNetworkMessage(nm)
	
	nm2 := GetNetworkMessage()
	if nm2.Type != "" || len(nm2.Payload) != 0 {
		t.Error("NetworkMessage not properly cleared")
	}
	PutNetworkMessage(nm2)
}

// Benchmark allocation with pool vs without pool
func BenchmarkCharacterStateWithPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs := GetCharacterState()
		cs.ID = "test"
		cs.Health = 100
		PutCharacterState(cs)
	}
}

func BenchmarkCharacterStateWithoutPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs := &CharacterState{
			ID:     "test",
			Health: 100,
		}
		_ = cs // prevent optimization
	}
}

func BenchmarkAnimationFrameWithPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		af := GetAnimationFrame()
		af.ImageData = append(af.ImageData, 1, 2, 3)
		af.Width = 64
		PutAnimationFrame(af)
	}
}

func BenchmarkAnimationFrameWithoutPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		af := &AnimationFrame{
			ImageData: []byte{1, 2, 3},
			Width:     64,
		}
		_ = af // prevent optimization
	}
}