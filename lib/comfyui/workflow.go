package comfyui

// workflow.go implements a minimal workflow template loader with in‑memory
// caching and simple parameter substitution. This is the third incremental
// step of the ComfyUI integration (after HTTP submission & WebSocket progress)
// outlined in GIF_PLAN.md. The design intentionally favours a *boring* and
// easily testable approach:
//
//   * Standard library only: os, path/filepath, encoding/json, strings.
//   * Read‑through cache: templates parsed once then deep‑cloned on use.
//   * Placeholders: "{{SLOT}}" tokens inside string values are replaced
//     using a supplied map. We purposely restrict substitution to simple
//     string values (no expression language) to avoid complexity and security
//     pitfalls. Unknown placeholders are left intact so callers can detect
//     or perform multi‑phase substitution later.
//   * Concurrency safety: RWMutex protects cache map; each Instantiate call
//     performs a deep copy so callers may mutate returned workflow.
//   * Context awareness: Load / Instantiate respect ctx cancellation for
//     early exit during file IO or CPU bound cloning.
//   * Small functions (<30 LOC) with clear, wrapped errors.
//
// Future extensions (not in this slice):
//   * Template discovery directory scanning.
//   * Advanced validation of node graph schema.
//   * Parameter type coercion & numeric substitution.
//   * Partial evaluation / multi‑stage substitution.

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
)

// TemplateLoader loads and instantiates workflow templates from JSON files.
// It caches decoded templates keyed by absolute path to avoid repeated disk
// IO and JSON parsing. The loader is safe for concurrent use.
type TemplateLoader struct {
    mu    sync.RWMutex
    cache map[string]*Workflow // stored canonical templates; never mutated after insert
    // now function facilitates deterministic testing (can be overridden in tests).
    now func() time.Time
}

// NewTemplateLoader creates a new loader instance.
func NewTemplateLoader() *TemplateLoader {
    return &TemplateLoader{cache: make(map[string]*Workflow), now: time.Now}
}

// Load reads a workflow template JSON file from path. If the file was
// previously loaded it returns the cached instance (a deep copy is NOT made
// here; callers wanting a mutable copy should use Instantiate which both
// applies substitutions and clones). Load is provided primarily to allow
// callers to validate availability up front or inspect raw structure.
func (l *TemplateLoader) Load(ctx context.Context, path string) (*Workflow, error) {
    if path == "" {
        return nil, errors.New("path required")
    }
    abs, err := filepath.Abs(path)
    if err != nil {
        return nil, fmt.Errorf("abs path: %w", err)
    }
    // Fast path: cached.
    l.mu.RLock()
    wf, ok := l.cache[abs]
    l.mu.RUnlock()
    if ok {
        return wf, nil
    }
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    f, err := os.Open(abs)
    if err != nil {
        return nil, fmt.Errorf("open template: %w", err)
    }
    defer f.Close()
    data, err := io.ReadAll(f)
    if err != nil {
        return nil, fmt.Errorf("read template: %w", err)
    }
    var tmpl Workflow
    if err := json.Unmarshal(data, &tmpl); err != nil {
        return nil, fmt.Errorf("decode template: %w", err)
    }
    // Basic sanity: nodes map required (empty allowed but must exist).
    if tmpl.Nodes == nil {
        tmpl.Nodes = map[string]interface{}{}
    }
    l.mu.Lock()
    l.cache[abs] = &tmpl
    l.mu.Unlock()
    return &tmpl, nil
}

// Instantiate returns a deep copy of the cached (or freshly loaded) template
// with placeholder substitutions applied. The params map values replace
// occurrences of "{{KEY}}" inside any string fields within the workflow's
// JSON structure.
func (l *TemplateLoader) Instantiate(ctx context.Context, path string, params map[string]string) (*Workflow, error) {
    base, err := l.Load(ctx, path)
    if err != nil {
        return nil, err
    }
    // Deep clone via JSON round trip (simplest & safe for arbitrary map[string]interface{} graph).
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    raw, err := json.Marshal(base)
    if err != nil {
        return nil, fmt.Errorf("marshal base: %w", err)
    }
    var clone Workflow
    if err := json.Unmarshal(raw, &clone); err != nil {
        return nil, fmt.Errorf("unmarshal clone: %w", err)
    }
    if params != nil {
        applySubstitutions(&clone, params)
    }
    if clone.ID == "" { // ensure some ID for tracking; generate minimal fallback
        clone.ID = fmt.Sprintf("tmpl:%d", l.now().UnixNano())
    }
    return &clone, nil
}

// applySubstitutions walks the workflow structure and performs simple
// string placeholder replacement. Because the workflow structure is loosely
// typed (map[string]interface{}) we implement a small reflective style walker
// without using reflection — relying on type assertions only. This keeps the
// logic understandable and under 30 LOC.
func applySubstitutions(wf *Workflow, params map[string]string) {
    // Helper performs substitution on any string.
    replace := func(s string) string {
        if !strings.Contains(s, "{{") { // quick reject
            return s
        }
        for k, v := range params {
            token := "{{" + k + "}}"
            if strings.Contains(s, token) {
                s = strings.ReplaceAll(s, token, v)
            }
        }
        return s
    }
    // Walk nodes.
    for key, val := range wf.Nodes {
        wf.Nodes[key] = substituteValue(val, replace)
    }
    // Walk meta.
    for key, val := range wf.Meta {
        wf.Meta[key] = substituteValue(val, replace)
    }
}

// substituteValue recursively traverses composite JSON types performing
// string replacement where applicable.
func substituteValue(v interface{}, replace func(string) string) interface{} {
    switch t := v.(type) {
    case string:
        return replace(t)
    case []interface{}:
        for i, elem := range t {
            t[i] = substituteValue(elem, replace)
        }
        return t
    case map[string]interface{}:
        for k, vv := range t {
            t[k] = substituteValue(vv, replace)
        }
        return t
    default:
        return v
    }
}
