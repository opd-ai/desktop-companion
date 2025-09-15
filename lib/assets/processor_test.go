package assets

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"testing"
)

func createTestPNG(path string, w, h int) error {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func TestGIFAssembler_Process_Success(t *testing.T) {
	dir := t.TempDir()
	frames := []string{}
	for i := 0; i < 4; i++ {
		p := fmt.Sprintf("%s/frame%d.png", dir, i)
		if err := createTestPNG(p, 64, 64); err != nil {
			t.Fatalf("create png: %v", err)
		}
		frames = append(frames, p)
	}
	outPath := dir + "/out.gif"
	cfg := GIFConfig{Width: 64, Height: 64, FrameCount: 4, FrameRate: 10, MaxFileSize: 500000, Transparency: true}
	proc := &GIFAssembler{}
	err := proc.Process(context.Background(), frames, outPath, cfg)
	if err != nil {
		t.Errorf("expected success, got %v", err)
	}
	stat, err := os.Stat(outPath)
	if err != nil || stat.Size() == 0 {
		t.Errorf("gif not created or empty")
	}
}

func TestGIFAssembler_Process_NotEnoughFrames(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/frame.png"
	if err := createTestPNG(p, 64, 64); err != nil {
		t.Fatalf("create png: %v", err)
	}
	proc := &GIFAssembler{}
	cfg := GIFConfig{Width: 64, Height: 64, FrameCount: 4, FrameRate: 10, MaxFileSize: 500000, Transparency: true}
	err := proc.Process(context.Background(), []string{p}, dir+"/out.gif", cfg)
	if err == nil || err.Error() != "not enough frames: got 1, want 4" {
		t.Errorf("expected not enough frames error, got %v", err)
	}
}

func TestGIFAssembler_Process_BadPNG(t *testing.T) {
	dir := t.TempDir()
	bad := dir + "/bad.png"
	os.WriteFile(bad, []byte("not a png"), 0644)
	proc := &GIFAssembler{}
	cfg := GIFConfig{Width: 64, Height: 64, FrameCount: 1, FrameRate: 10, MaxFileSize: 500000, Transparency: true}
	err := proc.Process(context.Background(), []string{bad}, dir+"/out.gif", cfg)
	if err == nil || err.Error() == "" {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestGIFAssembler_Process_ContextCancel(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/frame.png"
	if err := createTestPNG(p, 64, 64); err != nil {
		t.Fatalf("create png: %v", err)
	}
	proc := &GIFAssembler{}
	cfg := GIFConfig{Width: 64, Height: 64, FrameCount: 1, FrameRate: 10, MaxFileSize: 500000, Transparency: true}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := proc.Process(ctx, []string{p}, dir+"/out.gif", cfg)
	if err == nil {
		t.Errorf("expected context canceled error, got %v", err)
	}
}
