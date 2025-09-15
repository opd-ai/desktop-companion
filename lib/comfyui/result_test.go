package comfyui

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

// helper to build client with custom handler for result endpoint only.
func newResultClient(t *testing.T, handler http.HandlerFunc) (Client, *httptest.Server) {
    srv := httptest.NewServer(handler)
    cfg := DefaultConfig()
    cfg.ServerURL = srv.URL
    cli, err := New(cfg)
    if err != nil { t.Fatalf("new client: %v", err) }
    return cli, srv
}

func TestGetResultSuccess(t *testing.T) {
    pngData := []byte{0x89, 0x50, 0x4E, 0x47}
    b64 := base64.StdEncoding.EncodeToString(pngData)
    handler := func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.URL.Path, "/api/results/") { t.Fatalf("unexpected path: %s", r.URL.Path) }
        _ = json.NewEncoder(w).Encode(map[string]interface{}{ 
            "job_id":"job1",
            "status":"completed",
            "artifacts": []map[string]string{{"filename":"image_0.png","mime":"image/png","b64":b64}},
        })
    }
    cli, srv := newResultClient(t, handler)
    defer srv.Close()
    res, err := cli.GetResult(context.Background(), "job1")
    if err != nil { t.Fatalf("get result: %v", err) }
    if len(res.Artifacts) != 1 || res.Artifacts[0].Filename != "image_0.png" { t.Fatalf("unexpected artifacts: %+v", res.Artifacts) }
    if string(res.Artifacts[0].Data) != string(pngData) { t.Fatalf("data mismatch") }
}

func TestGetResultEmptyJobID(t *testing.T) {
    cli, srv := newResultClient(t, func(w http.ResponseWriter, r *http.Request){})
    defer srv.Close()
    _, err := cli.GetResult(context.Background(), "")
    if err == nil || !strings.Contains(err.Error(), "jobID required") { t.Fatalf("expected jobID required, got %v", err) }
}

func TestGetResultNon200(t *testing.T) {
    cli, srv := newResultClient(t, func(w http.ResponseWriter, r *http.Request){ w.WriteHeader(http.StatusBadGateway); _, _ = w.Write([]byte("fail")) })
    defer srv.Close()
    _, err := cli.GetResult(context.Background(), "jobx")
    if err == nil || !strings.Contains(err.Error(), "unexpected status") { t.Fatalf("expected status error, got %v", err) }
}

func TestGetResultInvalidJSON(t *testing.T) {
    cli, srv := newResultClient(t, func(w http.ResponseWriter, r *http.Request){ _, _ = w.Write([]byte("{")) })
    defer srv.Close()
    _, err := cli.GetResult(context.Background(), "joby")
    if err == nil || !strings.Contains(err.Error(), "decode result json") { t.Fatalf("expected decode error, got %v", err) }
}

func TestGetResultInvalidBase64(t *testing.T) {
    cli, srv := newResultClient(t, func(w http.ResponseWriter, r *http.Request){
        _ = json.NewEncoder(w).Encode(map[string]interface{}{
            "job_id":"jobz","status":"completed","artifacts": []map[string]string{{"filename":"f.png","mime":"image/png","b64":"!!!"}},
        })
    })
    defer srv.Close()
    _, err := cli.GetResult(context.Background(), "jobz")
    if err == nil || !strings.Contains(err.Error(), "decode artifact") { t.Fatalf("expected base64 error, got %v", err) }
}

func TestSaveArtifactsSuccess(t *testing.T) {
    dir := t.TempDir()
    res := &JobResult{JobID:"a", Status:"completed", Artifacts: []Artifact{{Filename:"a.png", Data: []byte{1,2,3}, MIME:"image/png"}}}
    if err := SaveArtifacts(res, dir); err != nil { t.Fatalf("save: %v", err) }
    b, err := os.ReadFile(filepath.Join(dir, "a.png"))
    if err != nil { t.Fatalf("read: %v", err) }
    if len(b) != 3 || b[0] != 1 { t.Fatalf("unexpected file contents: %v", b) }
}

func TestSaveArtifactsCollision(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "a.png")
    if err := os.WriteFile(path, []byte("x"), 0o644); err != nil { t.Fatalf("prep: %v", err) }
    res := &JobResult{JobID:"a", Status:"completed", Artifacts: []Artifact{{Filename:"a.png", Data: []byte{1}}}}
    err := SaveArtifacts(res, dir)
    if err == nil || !strings.Contains(err.Error(), "file exists") { t.Fatalf("expected collision error, got %v", err) }
}

func TestSaveArtifactsBadDir(t *testing.T) {
    res := &JobResult{JobID:"a", Status:"completed", Artifacts: []Artifact{}}
    err := SaveArtifacts(res, "")
    if err == nil || !strings.Contains(err.Error(), "dir required") { t.Fatalf("expected dir required, got %v", err) }
}

func TestGetResultContextCancel(t *testing.T) {
    cli, srv := newResultClient(t, func(w http.ResponseWriter, r *http.Request){
        time.Sleep(30 * time.Millisecond)
        _, _ = w.Write([]byte(`{"job_id":"slow","status":"completed","artifacts":[]}`))
    })
    defer srv.Close()
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
    defer cancel()
    _, err := cli.GetResult(ctx, "slow")
    if err == nil || !strings.Contains(err.Error(), "context deadline") { t.Fatalf("expected context deadline, got %v", err) }
}
