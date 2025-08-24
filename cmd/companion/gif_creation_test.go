package main

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"testing"
)

// TestValidGIFCreation tests creating a valid minimal GIF using Go's standard library
func TestValidGIFCreation(t *testing.T) {
	// Create a simple 1x1 image
	img := image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{
		color.RGBA{255, 255, 255, 255}, // white
		color.RGBA{0, 0, 0, 255},       // black
	})
	img.SetColorIndex(0, 0, 0) // Set pixel to white

	// Create GIF
	var buf bytes.Buffer
	err := gif.Encode(&buf, img, nil)
	if err != nil {
		t.Fatalf("Failed to encode GIF: %v", err)
	}

	// Test decoding
	_, err = gif.DecodeAll(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to decode created GIF: %v", err)
	}

	t.Logf("Valid GIF created with %d bytes", buf.Len())
	t.Logf("GIF bytes: %v", buf.Bytes())
}
