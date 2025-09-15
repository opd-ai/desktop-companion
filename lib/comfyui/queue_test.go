package comfyui

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// fakeClient implements Client subset used for submission tests.
type fakeClient struct {
	submitDelay time.Duration
	active      int32
	maxObserved int32
	fail        bool
}

func (f *fakeClient) SubmitWorkflow(ctx context.Context, wf *Workflow) (*Job, error) {
	if f.fail {
		return nil, errors.New("forced error")
	}
	a := atomic.AddInt32(&f.active, 1)
	defer func() {
		atomic.AddInt32(&f.active, -1)
	}()
	for {
		// record max observed concurrency
		for {
			m := atomic.LoadInt32(&f.maxObserved)
			if a <= m {
				break
			}
			if atomic.CompareAndSwapInt32(&f.maxObserved, m, a) {
				break
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(f.submitDelay):
			return &Job{ID: wf.ID, Status: "queued"}, nil
		}
	}
}
func (f *fakeClient) GetQueueStatus(ctx context.Context) (*QueueStatus, error) {
	return &QueueStatus{}, nil
}
func (f *fakeClient) MonitorJob(ctx context.Context, jobID string) (<-chan JobProgress, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeClient) GetResult(ctx context.Context, jobID string) (*JobResult, error) {
	return nil, errors.New("not implemented")
}

func TestQueueManager_ConcurrencyLimit(t *testing.T) {
	fc := &fakeClient{submitDelay: 30 * time.Millisecond}
	qm := NewQueueManager(fc, 2)
	ctx := context.Background()
	wfs := []*Workflow{{ID: "a", Nodes: map[string]interface{}{}}, {ID: "b", Nodes: map[string]interface{}{}}, {ID: "c", Nodes: map[string]interface{}{}}}
	start := time.Now()
	jobs, err := qm.SubmitBatch(ctx, wfs)
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if len(jobs) != 3 {
		t.Fatalf("expected 3 jobs")
	}
	if atomic.LoadInt32(&fc.maxObserved) > 2 {
		t.Fatalf("expected max concurrency <=2, got %d", fc.maxObserved)
	}
	if dur := time.Since(start); dur < 60*time.Millisecond {
		t.Fatalf("expected at least ~2 intervals, got %v", dur)
	}
}

func TestQueueManager_SubmitContextCancel(t *testing.T) {
	fc := &fakeClient{submitDelay: 50 * time.Millisecond}
	qm := NewQueueManager(fc, 1)
	ctx, cancel := context.WithCancel(context.Background())
	// First submission occupies the slot.
	done1 := make(chan struct{})
	go func() { _, _ = qm.Submit(ctx, &Workflow{ID: "x", Nodes: map[string]interface{}{}}); close(done1) }()
	time.Sleep(10 * time.Millisecond)
	// Second submission should block then be canceled.
	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel2()
	_, err := qm.Submit(ctx2, &Workflow{ID: "y", Nodes: map[string]interface{}{}})
	if err == nil {
		t.Fatalf("expected context error")
	}
	cancel()
	<-done1
}

func TestQueueManager_BatchNilWorkflow(t *testing.T) {
	fc := &fakeClient{submitDelay: 5 * time.Millisecond}
	qm := NewQueueManager(fc, 2)
	ctx := context.Background()
	_, err := qm.SubmitBatch(ctx, []*Workflow{{ID: "a", Nodes: map[string]interface{}{}}, nil})
	if err == nil || err.Error() != "nil workflow in batch" {
		t.Fatalf("expected nil workflow error, got %v", err)
	}
}

func TestQueueManager_LimitNormalization(t *testing.T) {
	fc := &fakeClient{submitDelay: 1 * time.Millisecond}
	qm := NewQueueManager(fc, 0) // should normalize to 1
	ctx := context.Background()
	wfs := []*Workflow{{ID: "a", Nodes: map[string]interface{}{}}, {ID: "b", Nodes: map[string]interface{}{}}}
	_, err := qm.SubmitBatch(ctx, wfs)
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if atomic.LoadInt32(&fc.maxObserved) != 1 {
		t.Fatalf("expected concurrency 1, got %d", fc.maxObserved)
	}
}
