// connection_test.go: Unit tests for ConnectionManager recovery logic
// Simulates connection loss, recovery, and error scenarios

package network

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

type fakeConn struct{}

func (f *fakeConn) Read(b []byte) (int, error)         { return 0, nil }
func (f *fakeConn) Write(b []byte) (int, error)        { return 0, nil }
func (f *fakeConn) Close() error                       { return nil }
func (f *fakeConn) LocalAddr() net.Addr                { return nil }
func (f *fakeConn) RemoteAddr() net.Addr               { return nil }
func (f *fakeConn) SetDeadline(t time.Time) error      { return nil }
func (f *fakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (f *fakeConn) SetWriteDeadline(t time.Time) error { return nil }

func TestConnectSuccess(t *testing.T) {
	cm := &ConnectionManager{
		DialFunc: func(ctx context.Context) (net.Conn, error) {
			return &fakeConn{}, nil
		},
		MaxRetries: 2,
		BaseDelay:  10 * time.Millisecond,
	}
	ctx := context.Background()
	if err := cm.Connect(ctx); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !cm.IsConnected() {
		t.Errorf("expected connected")
	}
}

func TestConnectFailure(t *testing.T) {
	cm := &ConnectionManager{
		DialFunc: func(ctx context.Context) (net.Conn, error) {
			return nil, errors.New("fail")
		},
		MaxRetries: 2,
		BaseDelay:  1 * time.Millisecond,
	}
	ctx := context.Background()
	if err := cm.Connect(ctx); err == nil {
		t.Fatalf("expected error, got nil")
	}
	if cm.IsConnected() {
		t.Errorf("expected not connected")
	}
}

func TestConnectCancel(t *testing.T) {
	cm := &ConnectionManager{
		DialFunc: func(ctx context.Context) (net.Conn, error) {
			return nil, errors.New("fail")
		},
		MaxRetries: 5,
		BaseDelay:  1 * time.Millisecond,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	if err := cm.Connect(ctx); err == nil {
		t.Fatalf("expected cancel error, got nil")
	}
}
