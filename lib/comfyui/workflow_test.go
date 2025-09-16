package comfyui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// helper to create temp workflow file
func writeTempWorkflow(t *testing.T, dir, name, json string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(json), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return path
}

func TestTemplateLoader_LoadAndCache(t *testing.T) {
	dir := t.TempDir()
	path := writeTempWorkflow(t, dir, "tmpl.json", `{"id":"base","nodes":{"a":{"prompt":"hello"}},"meta":{"author":"me"}}`)
	l := NewTemplateLoader()
	ctx := context.Background()
	wf1, err := l.Load(ctx, path)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	wf2, err := l.Load(ctx, path)
	if err != nil {
		t.Fatalf("second load: %v", err)
	}
	if wf1 != wf2 {
		t.Fatalf("expected cached pointer equality")
	}
}

func TestTemplateLoader_MissingFile(t *testing.T) {
	l := NewTemplateLoader()
	_, err := l.Load(context.Background(), "nope.json")
	if err == nil || !strings.Contains(err.Error(), "open template") {
		t.Fatalf("expected open template error, got %v", err)
	}
}

func TestTemplateLoader_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := writeTempWorkflow(t, dir, "bad.json", `{"id":123}`) // id numeric invalid for string
	l := NewTemplateLoader()
	_, err := l.Load(context.Background(), path)
	if err == nil || !strings.Contains(err.Error(), "decode template") {
		t.Fatalf("expected decode template error, got %v", err)
	}
}

func TestTemplateLoader_InstantiateSubstitution(t *testing.T) {
	dir := t.TempDir()
	json := `{"nodes":{"n1":{"text":"Value {{FOO}} and {{BAR}}"},"n2":["{{FOO}}", {"inner":"{{BAR}}"}]},"meta":{"desc":"{{FOO}} only"}}`
	path := writeTempWorkflow(t, dir, "sub.json", json)
	l := NewTemplateLoader()
	ctx := context.Background()
	wf, err := l.Instantiate(ctx, path, map[string]string{"FOO": "X", "BAR": "Y"})
	if err != nil {
		t.Fatalf("instantiate: %v", err)
	}
	got := wf.Nodes["n1"].(map[string]interface{})["text"].(string)
	if got != "Value X and Y" {
		t.Fatalf("substitution failed: %s", got)
	}
	arr := wf.Nodes["n2"].([]interface{})
	if arr[0].(string) != "X" {
		t.Fatalf("array substitution failed: %v", arr[0])
	}
	inner := arr[1].(map[string]interface{})["inner"].(string)
	if inner != "Y" {
		t.Fatalf("nested substitution failed: %s", inner)
	}
	if wf.Meta["desc"].(string) != "X only" {
		t.Fatalf("meta substitution failed: %v", wf.Meta["desc"])
	}
}

func TestTemplateLoader_InstantiateDeepCopy(t *testing.T) {
	dir := t.TempDir()
	path := writeTempWorkflow(t, dir, "deep.json", `{"nodes":{"a":{"v":"1"}}}`)
	l := NewTemplateLoader()
	ctx := context.Background()
	wf1, err := l.Instantiate(ctx, path, nil)
	if err != nil {
		t.Fatalf("instantiate1: %v", err)
	}
	wf2, err := l.Instantiate(ctx, path, nil)
	if err != nil {
		t.Fatalf("instantiate2: %v", err)
	}
	if wf1 == wf2 {
		t.Fatalf("expected different instances")
	}
	// mutate wf1 and ensure wf2 unchanged
	wf1.Nodes["a"].(map[string]interface{})["v"] = "changed"
	if wf2.Nodes["a"].(map[string]interface{})["v"].(string) != "1" {
		t.Fatalf("mutation leaked to second instance")
	}
}

func TestTemplateLoader_ContextCancel(t *testing.T) {
	// Use a large file to give a window for cancellation; replicate content.
	dir := t.TempDir()
	builder := strings.Builder{}
	builder.WriteString(`{"nodes":{"n":"`) // start JSON
	for i := 0; i < 20000; i++ {           // ~some KB
		builder.WriteString("data")
	}
	builder.WriteString(`"}}`)
	path := writeTempWorkflow(t, dir, "big.json", builder.String())
	l := NewTemplateLoader()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	_, err := l.Instantiate(ctx, path, nil)
	if err == nil {
		t.Fatalf("expected context error")
	}
	if !strings.Contains(err.Error(), "context canceled") && !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("expected cancellation error, got %v", err)
	}
}

func TestTemplateLoader_GeneratedID(t *testing.T) {
	dir := t.TempDir()
	path := writeTempWorkflow(t, dir, "noid.json", `{"nodes":{}}`)
	l := NewTemplateLoader()
	l.now = func() time.Time { return time.Unix(10, 5) }
	wf, err := l.Instantiate(context.Background(), path, nil)
	if err != nil {
		t.Fatalf("instantiate: %v", err)
	}
	if wf.ID != "tmpl:10000000005" {
		t.Fatalf("unexpected generated id: %s", wf.ID)
	}
}
