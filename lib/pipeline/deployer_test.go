package pipeline

import (
	"image"
	"image/color"
	"image/gif"
	"os"
	"testing"
)

func createTestGIF(path string, frames int, transparent bool) error {
	imgs := []*image.Paletted{}
	delays := []int{}
	for i := 0; i < frames; i++ {
		pal := color.Palette{color.Black, color.White}
		if transparent {
			pal[0] = color.Transparent
		}
		img := image.NewPaletted(image.Rect(0, 0, 64, 64), pal)
		imgs = append(imgs, img)
		delays = append(delays, 10)
	}
	g := &gif.GIF{Image: imgs, Delay: delays}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, g)
}

func TestValidateGIF_Success(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/ok.gif"
	if err := createTestGIF(p, 4, true); err != nil {
		t.Fatalf("create gif: %v", err)
	}
	cfg := GIFConfig{FrameCount: 4, MaxFileSize: 500000, Transparency: true}
	if err := ValidateGIF(p, cfg); err != nil {
		t.Errorf("expected success, got %v", err)
	}
}

func TestValidateGIF_TooLarge(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/big.gif"
	if err := createTestGIF(p, 4, true); err != nil {
		t.Fatalf("create gif: %v", err)
	}
	cfg := GIFConfig{FrameCount: 4, MaxFileSize: 1, Transparency: true}
	if err := ValidateGIF(p, cfg); err == nil {
		t.Errorf("expected too large error")
	}
}

func TestValidateGIF_BadFrameCount(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/badframe.gif"
	if err := createTestGIF(p, 2, true); err != nil {
		t.Fatalf("create gif: %v", err)
	}
	cfg := GIFConfig{FrameCount: 4, MaxFileSize: 500000, Transparency: true}
	if err := ValidateGIF(p, cfg); err == nil {
		t.Errorf("expected bad frame count error")
	}
}

func TestValidateGIF_NoTransparency(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/notrans.gif"
	if err := createTestGIF(p, 4, false); err != nil {
		t.Fatalf("create gif: %v", err)
	}
	cfg := GIFConfig{FrameCount: 4, MaxFileSize: 500000, Transparency: true}
	if err := ValidateGIF(p, cfg); err == nil {
		t.Errorf("expected no transparency error")
	}
}
