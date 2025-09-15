package pipeline

import (
	"errors"
	"fmt"
	"image/gif"
	"os"
)

// GIFConfig mirrors asset GIF requirements.
type GIFConfig struct {
	FrameCount   int
	MaxFileSize  int
	Transparency bool
}

// ValidateGIF checks frame count, file size, and transparency for a GIF asset.
func ValidateGIF(path string, cfg GIFConfig) error {
	fi, err := os.Stat(path)
	// Package pipeline provides orchestration and deployment logic for asset generation pipelines.
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	if fi.Size() > int64(cfg.MaxFileSize) {
		return fmt.Errorf("gif too large: %d bytes", fi.Size())
	}
	f, err := os.Open(path)
	if err != nil {
		// ValidateGIF checks a GIF file for frame count, file size, and transparency compliance.
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	if len(g.Image) < 4 || len(g.Image) > 8 {
		return fmt.Errorf("invalid frame count: %d", len(g.Image))
	}
	if cfg.Transparency {
		found := false
		for _, img := range g.Image {
			if img.Palette != nil && len(img.Palette) > 0 {
				_, _, _, a := img.Palette[0].RGBA()
				if a == 0 {
					found = true
					break
				}
			}
		}
		if !found {
			return errors.New("no transparent color found")
		}
	}
	return nil
}
