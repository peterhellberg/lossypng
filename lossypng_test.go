package lossypng

import (
	"bufio"
	"bytes"
	"image"
	"image/png"
	"log"
	"os"
	"testing"
)

func TestOptimize(t *testing.T) {
	m, _ := openTestImage("_testdata/dakar.png")
	o := Optimize(m, RGBAConversion, 8)
	ob := pngBuffer(o)

	if got, max := ob.Len(), 179725; got > max {
		t.Errorf("ob.Len() = %d, expected max %d", got, max)
	}

	g := Optimize(m, GrayscaleConversion, 20)
	gb := pngBuffer(g)

	if got, max := gb.Len(), 26265; got > max {
		t.Errorf("gb.Len() = %d, expected max %d", got, max)
	}
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

func pngBuffer(m image.Image) bytes.Buffer {
	var b bytes.Buffer

	w := bufio.NewWriter(&b)

	png.Encode(w, m)

	w.Flush()

	return b
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
