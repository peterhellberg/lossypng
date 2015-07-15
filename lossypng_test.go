package lossypng

import (
	"image"
	"image/png"
	"log"
	"os"
	"testing"
)

func TestOptimize(t *testing.T) {
	m, _ := openTestImage("_testdata/dakar.png")
	o := Optimize(m, RGBAConversion, 10)
	g := Optimize(m, GrayscaleConversion, 10)

	savePNGImage(m, "/tmp/lossypng-original.png")
	savePNGImage(o, "/tmp/lossypng-optimized.png")
	savePNGImage(g, "/tmp/lossypng-grayscale.png")
}

func openTestImage(fn string) (image.Image, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, _, err := image.Decode(f)

	return m, err
}

// Helper method to debug processed image
func savePNGImage(m image.Image, fn string) {
	log.Printf("saving: %s", fn)
	if f, err := os.Create(fn); err != nil {
		log.Printf("unable to create %s: %v", fn, err)
	} else {
		defer f.Close()

		enc := png.Encoder{png.BestCompression}
		enc.Encode(f, m)
	}
}
