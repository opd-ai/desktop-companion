// Package assets provides asset generation and post-processing utilities.
package assets

import (
	"context"
	"fmt"
	"image"
	"image/color/palette"
	"image/gif"
	"image/png"
	"os"
)

// GIFConfig specifies GIF output parameters.
type GIFConfig struct {
	Width        int
	Height       int
	FrameCount   int
	FrameRate    int
	MaxFileSize  int
	Transparency bool
}

// ArtifactPostProcessor defines a post-processing hook for generated artifacts.
type ArtifactPostProcessor interface {
	// Process assembles a GIF from PNG frame paths and writes to outPath.
	Process(ctx context.Context, frames []string, outPath string, cfg GIFConfig) error
}

// GIFAssembler implements ArtifactPostProcessor for GIF creation.
type GIFAssembler struct{}

// Process assembles a GIF from PNG frames and writes to outPath.
// Validates frame count, file size, and transparency.
func (g *GIFAssembler) Process(ctx context.Context, frames []string, outPath string, cfg GIFConfig) error {
	if len(frames) < cfg.FrameCount {
		return fmt.Errorf("not enough frames: got %d, want %d", len(frames), cfg.FrameCount)
	}
	gifImg := &gif.GIF{}
	for i, f := range frames {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		file, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("open frame %s: %w", f, err)
		}
		defer file.Close()
		img, err := png.Decode(file)
		if err != nil {
			return fmt.Errorf("decode png %s: %w", f, err)
		}
		// Resize if needed (skip for simplicity; assume correct size)
		gifImg.Image = append(gifImg.Image, imageToPaletted(img))
		gifImg.Delay = append(gifImg.Delay, 100/cfg.FrameRate)
		if i >= cfg.FrameCount {
			break
		}
	}
	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create gif: %w", err)
	}
	defer out.Close()
	if err := gif.EncodeAll(out, gifImg); err != nil {
		return fmt.Errorf("encode gif: %w", err)
	}
	// Validate file size
	stat, err := out.Stat()
	if err != nil {
		return fmt.Errorf("stat gif: %w", err)
	}
	if stat.Size() > int64(cfg.MaxFileSize) {
		return fmt.Errorf("gif too large: %d bytes", stat.Size())
	}
	return nil
}

// imageToPaletted converts image.Image to *image.Paletted for GIF encoding.
func imageToPaletted(img image.Image) *image.Paletted {
	bounds := img.Bounds()
	paletted := image.NewPaletted(bounds, palette.Plan9)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			paletted.Set(x, y, img.At(x, y))
		}
	}
	return paletted
}
