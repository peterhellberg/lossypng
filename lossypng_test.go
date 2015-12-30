package lossypng

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
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

	g = Optimize(m, NoConversion, 20)
	gb = pngBuffer(g)

	if got, max := gb.Len(), 91380; got > max {
		t.Errorf("gb.Len() = %d, expected max %d", got, max)
	}
}

func TestPaethPredictor(t *testing.T) {
	for i, tt := range []struct {
		a, b, c, want uint8
	}{
		{1, 2, 3, 1},
		{4, 5, 6, 4},
		{7, 8, 9, 7},
		{2, 4, 1, 4},
	} {
		if got := paethPredictor(tt.a, tt.b, tt.c); got != tt.want {
			t.Fatalf(`[%d] paethPredictor(%d, %d, %d) = %d, want %d`,
				i, tt.a, tt.b, tt.c, got, tt.want)
		}
	}
}

func TestColorDifference(t *testing.T) {
	for i, tt := range []struct {
		a color.Color
		b color.Color
		d colorDelta
	}{
		{color.White, color.Black, colorDelta{10920, 87380, 10921, 0}},
		{color.Black, color.White, colorDelta{-10920, -87380, -10921, 0}},
		{color.White, color.RGBA{0xff, 0x66, 0x00, 0xff}, colorDelta{0, 52428, -1, 0}},
	} {
		d := colorDifference(tt.a, tt.b)
		if d[0] != tt.d[0] || d[1] != tt.d[1] || d[2] != tt.d[2] || d[3] != tt.d[3] {
			t.Fatalf("[%d] colorDifference(tt.a, tt.b) = %v, want %v", i, d, tt.d)
		}
	}
}

func BenchmarkOptimizeNoConversionQ1(b *testing.B) {
	m, _ := openTestImage("_testdata/dakar.png")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Optimize(m, NoConversion, 1)
	}
}

func BenchmarkOptimizeNoConversionQ10(b *testing.B) {
	m, _ := openTestImage("_testdata/dakar.png")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Optimize(m, NoConversion, 10)
	}
}

func BenchmarkOptimizeNoConversionQ20(b *testing.B) {
	m, _ := openTestImage("_testdata/dakar.png")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Optimize(m, NoConversion, 20)
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

		enc := png.Encoder{CompressionLevel: png.BestCompression}
		enc.Encode(f, m)
	}
}
