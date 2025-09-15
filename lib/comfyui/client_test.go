package comfyui

import (
    "context"
    "errors"
    "fmt"
    "net/http"
    "net/http/httptest"
    "strings"
    "sync/atomic"
    "testing"
    "time"
    "encoding/json"
)

// Test configuration validation
func TestConfigValidate(t *testing.T) {
    c := DefaultConfig()
    if err := c.Validate(); err != nil { t.Fatalf("expected valid config: %v", err) }
    c.ServerURL = ""
    if err := c.Validate(); err == nil { t.Fatalf("expected error for empty server URL") }
}

// helper to create client with test server
func newTestClient(t *testing.T, handler http.HandlerFunc) (Client, *httptest.Server) {
    srv := httptest.NewServer(handler)
    cfg := DefaultConfig()
    cfg.ServerURL = srv.URL
    cfg.RetryAttempts = 1
    cfg.RetryBackoff = 10 * time.Millisecond
    cli, err := New(cfg)
    if err != nil { t.Fatalf("new client: %v", err) }
    return cli, srv
}

func TestSubmitWorkflowSuccess(t *testing.T) {
    cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/api/workflows") {
            t.Fatalf("unexpected path/method: %s %s", r.Method, r.URL.Path)
        }
        w.WriteHeader(http.StatusAccepted)
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte(`{"id":"job123","status":"queued"}`))
    })
    defer srv.Close()
    ctx := context.Background()
    job, err := cli.SubmitWorkflow(ctx, &Workflow{ID: "wf1", Nodes: map[string]interface{}{"n":1}})
    if err != nil { t.Fatalf("submit failed: %v", err) }
    if job.ID != "job123" { t.Fatalf("unexpected job id: %s", job.ID) }
}

func TestSubmitWorkflowServerErrorRetry(t *testing.T) {
    var calls int32
    cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
        c := atomic.AddInt32(&calls, 1)
        if c == 1 { // first call 500 triggers retry
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusAccepted)
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte(`{"id":"job456","status":"queued"}`))
    })
    defer srv.Close()
    job, err := cli.SubmitWorkflow(context.Background(), &Workflow{ID: "wf2", Nodes: map[string]interface{}{}})
    if err != nil { t.Fatalf("expected success after retry: %v", err) }
    if job.ID != "job456" { t.Fatalf("unexpected job id: %s", job.ID) }
    if atomic.LoadInt32(&calls) != 2 { t.Fatalf("expected 2 calls, got %d", calls) }
}

func TestSubmitWorkflowNonRetryable(t *testing.T) {
    cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
        _, _ = w.Write([]byte("bad request"))
    })
    defer srv.Close()
    _, err := cli.SubmitWorkflow(context.Background(), &Workflow{ID: "wf3", Nodes: map[string]interface{}{}})
    if err == nil || !strings.Contains(err.Error(), "unexpected status") { t.Fatalf("expected client error, got %v", err) }
}

func TestSubmitWorkflowContextCancel(t *testing.T) {
    cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(50 * time.Millisecond)
        w.WriteHeader(http.StatusAccepted)
        _, _ = w.Write([]byte(`{"id":"late","status":"queued"}`))
    })
    defer srv.Close()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
    defer cancel()
    _, err := cli.SubmitWorkflow(ctx, &Workflow{ID: "wf4", Nodes: map[string]interface{}{}})
    if err == nil { t.Fatalf("expected context error") }
    if !errors.Is(err, context.DeadlineExceeded) { t.Fatalf("expected deadline exceeded, got %v", err) }
}

func TestGetQueueStatusSuccess(t *testing.T) {
    cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/api/queue") {
            _ = json.NewEncoder(w).Encode(QueueStatus{Pending:2, Running:1, Finished:5})
            return
        }
        t.Fatalf("unexpected path %s", r.URL.Path)
    })
    defer srv.Close()
    qs, err := cli.GetQueueStatus(context.Background())
    if err != nil { t.Fatalf("get queue: %v", err) }
    if qs.Pending != 2 || qs.Running != 1 || qs.Finished !=5 { t.Fatalf("unexpected queue status: %+v", qs) }
}

func TestGetQueueStatusError(t *testing.T) {
    cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadGateway)
        _, _ = w.Write([]byte("upstream"))
    })
    defer srv.Close()
    _, err := cli.GetQueueStatus(context.Background())
    if err == nil || !strings.Contains(err.Error(), "unexpected status") { t.Fatalf("expected error, got %v", err) }
}

func TestSubmitWorkflowRetryExhaustion(t *testing.T) {
    var calls int32
    cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
        atomic.AddInt32(&calls, 1)
        w.WriteHeader(http.StatusInternalServerError)
    })
    defer srv.Close()
    _, err := cli.SubmitWorkflow(context.Background(), &Workflow{ID: "wf5", Nodes: map[string]interface{}{}})
    if err == nil || !strings.Contains(err.Error(), "server error") { t.Fatalf("expected server error, got %v", err) }
    if calls < 2 { t.Fatalf("expected at least 2 attempts, got %d", calls) }
}

func TestSubmitWorkflowInvalidJSON(t *testing.T) {
    cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusAccepted)
        _, _ = w.Write([]byte("not-json"))
    })
    defer srv.Close()
    _, err := cli.SubmitWorkflow(context.Background(), &Workflow{ID: "wf6", Nodes: map[string]interface{}{}})
    if err == nil || !strings.Contains(err.Error(), "decode job response") { t.Fatalf("expected decode error, got %v", err) }
}

func TestSubmitWorkflowEmptyJobID(t *testing.T) {
    cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusAccepted)
        _, _ = w.Write([]byte(`{"status":"queued"}`))
    })
    defer srv.Close()
    _, err := cli.SubmitWorkflow(context.Background(), &Workflow{ID: "wf7", Nodes: map[string]interface{}{}})
    if err == nil || !strings.Contains(err.Error(), "empty job id") { t.Fatalf("expected empty id error, got %v", err) }
}

// Ensure isRetryable logic roughly matches expectations.
func TestIsRetryable(t *testing.T) {
    cases := []struct{ in error; want bool }{
        {fmt.Errorf("server error 500"), true},
        {fmt.Errorf("connection refused"), true},
        {fmt.Errorf("timeout awaiting response"), true},
        {fmt.Errorf("unexpected status 400"), false},
        {nil, false},
    }
    for i, cse := range cases {
        if got := isRetryable(cse.in); got != cse.want { t.Fatalf("case %d got %v want %v", i, got, cse.want) }
    }
}
