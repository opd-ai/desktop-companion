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
    ws "nhooyr.io/websocket"
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

// --- WebSocket Monitoring Tests ---

// startWSServer spins up a test HTTP server upgrading a single path to websocket.
func startWSServer(t *testing.T, handler func(ctx context.Context, c *ws.Conn)) *httptest.Server {
    t.Helper()
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.URL.RawQuery, "job_id=") {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        c, err := ws.Accept(w, r, nil)
        if err != nil { t.Fatalf("accept: %v", err) }
        // Use background context so writes are not cancelled when handler returns.
        go handler(context.Background(), c)
    }))
}

func TestMonitorJobSuccess(t *testing.T) {
    srv := startWSServer(t, func(ctx context.Context, c *ws.Conn) {
        msgs := []string{
            `{"job_id":"abc","status":"running","progress":0.25}`,
            `{"job_id":"abc","status":"running","progress":0.50}`,
            `{"job_id":"abc","status":"completed","progress":1}`,
        }
        for _, m := range msgs {
            if err := c.Write(ctx, ws.MessageText, []byte(m)); err != nil { return }
        }
        c.Close(ws.StatusNormalClosure, "done")
    })
    defer srv.Close()
    cfg := DefaultConfig(); cfg.ServerURL = srv.URL; cfg.WSPath = "/"
    cli, err := New(cfg)
    if err != nil { t.Fatalf("new: %v", err) }
    ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
    defer cancel()
    ch, err := cli.MonitorJob(ctx, "abc")
    if err != nil { t.Fatalf("monitor: %v", err) }
    var count int
    for range ch { count++ }
    if count != 3 { t.Fatalf("expected 3 progress messages, got %d", count) }
}

func TestMonitorJobMalformedJSON(t *testing.T) {
    srv := startWSServer(t, func(ctx context.Context, c *ws.Conn) {
        _ = c.Write(ctx, ws.MessageText, []byte(`{"job_id":"jjj","status":"running","progress":0.1}`))
        _ = c.Write(ctx, ws.MessageText, []byte("not-json"))
        time.Sleep(20 * time.Millisecond)
        c.Close(ws.StatusNormalClosure, "done")
    })
    defer srv.Close()
    cfg := DefaultConfig(); cfg.ServerURL = srv.URL; cfg.WSPath = "/"
    cli, _ := New(cfg)
    ch, err := cli.MonitorJob(context.Background(), "jjj")
    if err != nil { t.Fatalf("monitor: %v", err) }
    var sawDecodeErr bool
    for p := range ch {
        if p.Err != nil && strings.Contains(p.Err.Error(), "decode progress") { sawDecodeErr = true }
    }
    if !sawDecodeErr { t.Fatalf("expected decode progress error") }
}

func TestMonitorJobContextCancel(t *testing.T) {
    srv := startWSServer(t, func(ctx context.Context, c *ws.Conn) {
        // continually send until client context cancels
        ticker := time.NewTicker(20 * time.Millisecond)
        defer ticker.Stop()
        for i := 0; i < 20; i++ { // max 400ms
            _ = c.Write(ctx, ws.MessageText, []byte(`{"job_id":"cxl","status":"running","progress":0.1}`))
            time.Sleep(20 * time.Millisecond)
        }
    })
    defer srv.Close()
    cfg := DefaultConfig(); cfg.ServerURL = srv.URL; cfg.WSPath = "/"
    cli, _ := New(cfg)
    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
    defer cancel()
    ch, err := cli.MonitorJob(ctx, "cxl")
    if err != nil { t.Fatalf("monitor: %v", err) }
    var sawCancel bool
    for p := range ch {
        if p.Status == "cancelled" { sawCancel = true }
    }
    if !sawCancel { t.Fatalf("expected cancelled status") }
}

func TestMonitorJobDialError(t *testing.T) {
    cfg := DefaultConfig(); cfg.ServerURL = "http://127.0.0.1:0"; cfg.WSPath = "/"
    cli, _ := New(cfg)
    _, err := cli.MonitorJob(context.Background(), "id1")
    if err == nil { t.Fatalf("expected dial error") }
}

func TestMonitorJobEmptyID(t *testing.T) {
    cfg := DefaultConfig(); cfg.ServerURL = "http://localhost:8188"; cfg.WSPath = "/"
    cli, _ := New(cfg)
    _, err := cli.MonitorJob(context.Background(), "")
    if err == nil || !strings.Contains(err.Error(), "jobID required") { t.Fatalf("expected jobID required error") }
}

// parseProgress internal helper tests
func TestParseProgressDecodeError(t *testing.T) {
    prog, terminal := parseProgress([]byte("not-json"), "fb")
    if prog.Err == nil || !terminal { t.Fatalf("expected decode error & terminal, got %+v term=%v", prog, terminal) }
}

func TestParseProgressTerminalState(t *testing.T) {
    data := []byte(`{"job_id":"x","status":"completed","progress":1}`)
    prog, terminal := parseProgress(data, "fallback")
    if prog.Err != nil || !terminal || prog.JobID != "x" { t.Fatalf("unexpected parse result %+v term=%v", prog, terminal) }
}

