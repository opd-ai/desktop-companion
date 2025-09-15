package comfyui

// queue.go provides a lightweight client-side submission queue with
// concurrency throttling. This is intentionally *not* a sophisticated job
// scheduler; ComfyUI itself maintains its own internal queue. We use this to
// prevent overwhelming the server with a sudden burst of submissions and to
// provide a single location for future enhancements (retry policies, local
// priority ordering, backoff, metrics hooks) without complicating the core
// Client interface.
//
// Design Principles:
//   * Standard library only - channel based semaphore for concurrency.
//   * Small, boring API - Submit (single) and SubmitBatch (slice) only.
//   * Non-invasive - wraps an existing Client; no interface changes needed.
//   * Context-aware - if the caller's context is cancelled while waiting for
//     a slot, the submission is aborted early.
//   * Order preservation in SubmitBatch - jobs returned in same order as
//     workflows slice (helpful for correlation at higher levels).
//   * Functions kept <30 LOC each for readability & testability.

import (
	"context"
	"errors"
	"fmt"
)

// QueueManager limits concurrent workflow submissions to the underlying
// ComfyUI Client. Safe for concurrent use.
type QueueManager struct {
	client Client
	sem    chan struct{}
}

// NewQueueManager creates a new QueueManager. If limit <= 0 it defaults to 1.
func NewQueueManager(client Client, limit int) *QueueManager {
	if limit <= 0 {
		limit = 1
	}
	return &QueueManager{client: client, sem: make(chan struct{}, limit)}
}

// Submit submits a single workflow respecting the concurrency limit. It
// returns the resulting Job or an error. Context cancellation while waiting
// for a slot or during submission returns promptly.
func (qm *QueueManager) Submit(ctx context.Context, wf *Workflow) (*Job, error) {
	if wf == nil {
		return nil, errors.New("workflow nil")
	}
	// Acquire slot or abort if context cancelled.
	select {
	case qm.sem <- struct{}{}:
		// acquired
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-qm.sem }()
	job, err := qm.client.SubmitWorkflow(ctx, wf)
	if err != nil {
		return nil, fmt.Errorf("submit workflow %s: %w", wf.ID, err)
	}
	return job, nil
}

// SubmitBatch submits multiple workflows concurrently (bounded by the queue
// limit) and returns their Jobs in the same slice order. The first error
// encountered cancels remaining pending submissions via the provided context
// if it is a cancelable context (callers can wrap with context.WithCancel).
func (qm *QueueManager) SubmitBatch(ctx context.Context, wfs []*Workflow) ([]*Job, error) {
	jobs := make([]*Job, len(wfs))
	errCh := make(chan error, 1)
	type pair struct {
		idx int
		job *Job
	}
	resCh := make(chan pair, len(wfs))

	// Launch each submission in its own goroutine; bounded by semaphore in Submit.
	for i, wf := range wfs {
		idx, wf := i, wf
		go func() {
			if wf == nil {
				// Treat nil workflow as immediate error.
				select {
				case errCh <- errors.New("nil workflow in batch"):
				default:
				}
				return
			}
			job, err := qm.Submit(ctx, wf)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			resCh <- pair{idx: idx, job: job}
		}()
	}

	// Collect results.
	remaining := len(wfs)
	for remaining > 0 {
		select {
		case err := <-errCh:
			return nil, err
		case p := <-resCh:
			jobs[p.idx] = p.job
			remaining--
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return jobs, nil
}
