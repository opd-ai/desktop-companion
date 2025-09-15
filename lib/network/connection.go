// connection.go: ConnectionManager for multiplayer networking
// Handles connection lifecycle, error recovery, and reconnection with backoff
// Uses Go stdlib only; all network operations use interface types for testability

package network

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// ConnectionManager manages a network connection with recovery and backoff.
// All network operations use net.Conn interface for testability.
// Recovery logic: on error, attempts reconnect with exponential backoff up to MaxRetries.
// Errors are returned explicitly; no panics.
type ConnectionManager struct {
	DialFunc   func(ctx context.Context) (net.Conn, error) // injectable for testing
	MaxRetries int
	BaseDelay  time.Duration
	mu         sync.Mutex
	conn       net.Conn
}

// Connect establishes a connection, retrying on failure with exponential backoff.
// Returns error if all attempts fail or context is canceled.
func (cm *ConnectionManager) Connect(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	var lastErr error
	for i := 0; i <= cm.MaxRetries; i++ {
		conn, err := cm.DialFunc(ctx)
		if err == nil {
			cm.conn = conn
			return nil
		}
		lastErr = err
		delay := cm.BaseDelay << i // exponential backoff
		select {
		case <-ctx.Done():
			return fmt.Errorf("connect canceled: %w", ctx.Err())
		case <-time.After(delay):
		}
	}
	return fmt.Errorf("connect failed after %d retries: %w", cm.MaxRetries, lastErr)
}

// Close closes the current connection if active.
// Returns error from net.Conn.Close or nil if not connected.
func (cm *ConnectionManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.conn != nil {
		return cm.conn.Close()
	}
	return nil
}

// IsConnected returns true if a connection is active.
func (cm *ConnectionManager) IsConnected() bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.conn != nil
}
