package comfyui

// result.go provides job result retrieval & artifact persistence. This is the
// next incremental ComfyUI capability after workflow templates. Design goals:
//
//   * Standard library only (net/http already in client.go, plus encoding/base64)
//   * Small focused functions (<30 LOC) with explicit error wrapping.
//   * Simple assumed JSON schema so we can iterate without locking into a
//     complex contract prematurely. We model only the fields needed now.
//   * Decouple network retrieval (GetResult) from disk persistence
//     (SaveArtifacts) to keep single responsibility and test them separately.
//   * Defensive validation: reject empty job IDs, verify status codes, limit
//     read size for safety (basic cap), surface detailed context in errors.
//
// Assumed server JSON (minimal subset):
// {
//   "job_id": "abc",
//   "status": "completed",
//   "artifacts": [
//      {"filename":"image_0.png","mime":"image/png","b64":"..."},
//      {"filename":"meta.json","mime":"application/json","b64":"..."}
//   ]
// }
// Unknown fields are ignored. The server may omit artifacts for failed jobs.
// We decode base64 into []byte; callers decide how to interpret.

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

// Artifact represents one generated output file for a job.
type Artifact struct {
    Filename string `json:"filename"` // Original or suggested file name
    MIME     string `json:"mime"`     // MIME type (advisory)
    Data     []byte `json:"-"`        // Decoded binary data
    B64      string `json:"b64"`      // Base64 string (only used during decode)
}

// JobResult is the decoded result payload returned by GetResult.
type JobResult struct {
    JobID     string     `json:"job_id"`
    Status    string     `json:"status"`
    Artifacts []Artifact `json:"artifacts"`
}

// GetResult retrieves a job's final result JSON from the server and decodes
// any base64 artifacts. It does not persist data to disk. Returns an error if
// the jobID is empty, the HTTP status is not 200, the JSON cannot be decoded,
// or any artifact fails base64 decoding.
func (c *client) GetResult(ctx context.Context, jobID string) (*JobResult, error) {
    if jobID == "" { return nil, errors.New("jobID required") }
    url := strings.TrimRight(c.cfg.ServerURL, "/") + "/api/results/" + jobID
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil { return nil, fmt.Errorf("create request: %w", err) }
    if c.cfg.APIKey != "" { req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey) }
    resp, err := c.httpc.Do(req)
    if err != nil { return nil, fmt.Errorf("get result: %w", err) }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
        return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
    }
    // Limit read to a conservative size (e.g. 10MB) to avoid excessive memory usage.
    const maxResultSize = 10 << 20
    data, err := io.ReadAll(io.LimitReader(resp.Body, maxResultSize+1))
    if err != nil { return nil, fmt.Errorf("read result body: %w", err) }
    if len(data) > maxResultSize { return nil, errors.New("result body too large") }
    var res JobResult
    if err := json.Unmarshal(data, &res); err != nil {
        return nil, fmt.Errorf("decode result json: %w", err)
    }
    // Basic sanity.
    if res.JobID == "" { res.JobID = jobID }
    for i := range res.Artifacts {
        if res.Artifacts[i].Filename == "" {
            return nil, fmt.Errorf("artifact %d missing filename", i)
        }
        if res.Artifacts[i].B64 != "" {
            decoded, err := base64.StdEncoding.DecodeString(res.Artifacts[i].B64)
            if err != nil { return nil, fmt.Errorf("decode artifact %s: %w", res.Artifacts[i].Filename, err) }
            res.Artifacts[i].Data = decoded
            // Clear B64 to avoid accidental re-encoding when marshalled.
            res.Artifacts[i].B64 = ""
        }
    }
    return &res, nil
}

// SaveArtifacts writes all artifacts to the specified directory. Filenames
// are sanitised (basic: strip path separators) to prevent directory traversal.
// If a file already exists, an error is returned (callers may remove or choose
// a different directory before retrying). Returns the first error encountered.
func SaveArtifacts(res *JobResult, dir string) error {
    if res == nil { return errors.New("result nil") }
    if dir == "" { return errors.New("dir required") }
    info, err := os.Stat(dir)
    if err != nil {
        if os.IsNotExist(err) {
            if err := os.MkdirAll(dir, 0o755); err != nil { return fmt.Errorf("mkdir: %w", err) }
        } else { return fmt.Errorf("stat dir: %w", err) }
    } else if !info.IsDir() { return fmt.Errorf("%s is not a directory", dir) }
    for _, a := range res.Artifacts {
        name := sanitizeFilename(a.Filename)
        if name == "" { return fmt.Errorf("empty sanitized filename for %q", a.Filename) }
        path := filepath.Join(dir, name)
        if _, err := os.Stat(path); err == nil { return fmt.Errorf("file exists: %s", name) }
        if err := os.WriteFile(path, a.Data, 0o644); err != nil { return fmt.Errorf("write %s: %w", name, err) }
    }
    return nil
}

// sanitizeFilename removes any directory components & simplistic dangerous elements.
func sanitizeFilename(fn string) string {
    fn = strings.TrimSpace(fn)
    fn = strings.ReplaceAll(fn, "\\", "/")
    if i := strings.LastIndex(fn, "/"); i >= 0 { fn = fn[i+1:] }
    fn = strings.Trim(fn, ". ")
    return fn
}
