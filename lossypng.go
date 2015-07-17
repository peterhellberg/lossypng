/*

Package lossypng is a library version of the lossypng command line tool.

Installation

    go get -u github.com/peterhellberg/lossypng

*/
package lossypng

import (
	"image"
	"image/color"
	"image/draw"
)

const (
	// NoConversion does not convert the image
	NoConversion = iota

	// GrayscaleConversion convert image to grayscale
	GrayscaleConversion

	// RGBAConversion convert image to 32-bit color
	RGBAConversion
)

const deltaComponents = 4

type colorDelta [deltaComponents]int32 // difference between two colors in rgba

// Optimize optimizes an image using the selected color conversion and quantization
func Optimize(original image.Image, colorConversion int, quantization int) image.Image {
	var (
		bounds    = original.Bounds()
		optimized = original // update optimized variable later if color conversion is necessary
	)

	switch colorConversion {
	case GrayscaleConversion:
		converted := image.NewGray(bounds)
		draw.Draw(converted, bounds, original, image.ZP, draw.Src)
		optimizeForAverageFilter(converted.Pix, bounds, converted.Stride, 1, quantization)
		optimized = converted
	case RGBAConversion:
		converted := image.NewRGBA(bounds)
		draw.Draw(converted, bounds, original, image.ZP, draw.Src)
		optimizeForAverageFilter(converted.Pix, bounds, converted.Stride, 4, quantization)
		optimized = converted
	default:
		// no color conversion requested
		switch o := original.(type) {
		case *image.Alpha:
			optimizeForAverageFilter(o.Pix, bounds, o.Stride, 1, quantization)
		case *image.Gray:
			optimizeForAverageFilter(o.Pix, bounds, o.Stride, 1, quantization)
		case *image.NRGBA:
			optimizeForAverageFilter(o.Pix, bounds, o.Stride, 4, quantization)
		case *image.Paletted:
			// many PNGs decode as image.Paletted
			// use alternative paeth optimizer for paletted images
			optimizeForPaethFilter(o.Pix, bounds, o.Stride, quantization, o.Palette)
		case *image.Alpha16:
			converted := image.NewAlpha(bounds)
			draw.Draw(converted, bounds, original, image.ZP, draw.Src)
			optimizeForAverageFilter(converted.Pix, bounds, converted.Stride, 1, quantization)
			optimized = converted
		case *image.Gray16:
			converted := image.NewGray(bounds)
			draw.Draw(converted, bounds, original, image.ZP, draw.Src)
			optimizeForAverageFilter(converted.Pix, bounds, converted.Stride, 1, quantization)
			optimized = converted
		default:
			// convert all other formats to RGBA
			// most JPEGs decode as image.YCbCr
			// most PNGs decode as image.RGBA
			converted := image.NewNRGBA(bounds)
			draw.Draw(converted, bounds, original, image.ZP, draw.Src)
			optimizeForAverageFilter(converted.Pix, bounds, converted.Stride, 4, quantization)
			optimized = converted
		}
	}

	return optimized
}

func optimizeForAverageFilter(pixels []uint8, bounds image.Rectangle, stride, bytesPerPixel int, quantization int) {
	if quantization < 1 {
		// Algorithm requires positive number.
		// Zero (or less) means lossless operation, so leaving image unchanged is correct.
		return
	}

	var (
		halfStep = int32(quantization / 2)
		height   = bounds.Dy()
		width    = bounds.Dx()
	)

	const (
		errorRowCount = 3
		filterWidth   = 5
		filterCenter  = 2
	)

	var colorError [errorRowCount][]colorDelta
	for i := 0; i < errorRowCount; i++ {
		colorError[i] = make([]colorDelta, width+filterWidth-1)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			diffusion := diffuseColorDeltas(colorError, x+filterCenter)
			for c := 0; c < bytesPerPixel; c++ {
				offset := y*stride + x*bytesPerPixel + c
				here := int32(pixels[offset])
				var errorHere int32
				if here > 0 && here < 255 {
					var up, left int32
					if y > 0 {
						up = int32(pixels[offset-stride])
					}
					if x > 0 {
						left = int32(pixels[offset-bytesPerPixel])
					}
					average := (up + left) / 2 // PNG average filter

					newValue := diffusion[c] + here - average
					newValue += halfStep
					newValue -= newValue % int32(quantization)
					newValue += average
					if newValue >= 0 && newValue <= 255 {
						pixels[offset] = uint8(newValue)
						errorHere = here - newValue
					}
				}
				colorError[0][x+filterCenter][c] = errorHere
			}
		}
		for i := 0; i < errorRowCount; i++ {
			colorError[(i+1)%errorRowCount] = colorError[i]
		}
	}
}

func optimizeForPaethFilter(pixels []uint8, bounds image.Rectangle, stride int, quantization int, palette color.Palette) {
	colorCount := len(palette)
	if colorCount <= 0 {
		return
	}

	var (
		height = bounds.Dy()
		width  = bounds.Dx()
	)

	const (
		errorRowCount = 3
		filterWidth   = 5
		filterCenter  = 2
	)

	var colorError [errorRowCount][]colorDelta
	for i := 0; i < errorRowCount; i++ {
		colorError[i] = make([]colorDelta, width+filterWidth-1)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var (
				diffusion = diffuseColorDeltas(colorError, x+filterCenter)
				offset    = y*stride + x
				here      = pixels[offset]

				up, left, diagonal uint8
			)

			if y > 0 {
				up = pixels[offset-stride]
			}

			if x > 0 {
				left = pixels[offset-1]
			}

			if y > 0 && x > 0 {
				diagonal = pixels[offset-stride-1]
			}

			var (
				paeth     = paethPredictor(left, up, diagonal) // PNG Paeth filter
				bestDelta = colorDifference(palette[here], palette[paeth])
				total     = bestDelta.add(diffusion)
				bestColor uint8
			)

			if (total.magnitude() >> 16) < uint64(quantization*quantization) {
				bestColor = paeth
			} else {
				bestDelta = colorDifference(palette[here], palette[bestColor])
				total = bestDelta.add(diffusion)
				bestMagnitude := total.magnitude()

				for i, candidate := range palette {
					delta := colorDifference(palette[here], candidate)
					total = delta.add(diffusion)
					nextMagnitude := total.magnitude()

					if bestMagnitude > nextMagnitude {
						bestMagnitude = nextMagnitude
						bestDelta = delta
						bestColor = uint8(i)
					}
				}
			}
			pixels[offset] = bestColor
			colorError[0][x+filterCenter] = bestDelta
		}
		for i := 0; i < errorRowCount; i++ {
			colorError[(i+1)%errorRowCount] = colorError[i]
		}
	}
}

// a = left, b = above, c = upper left
func paethPredictor(a, b, c uint8) uint8 {
	var (
		// Initial estimate
		p = int(a) + int(b) - int(c)

		// Distances to a, b, c
		pa = abs(p - int(a))
		pb = abs(p - int(b))
		pc = abs(p - int(c))
	)

	// Return nearest of a,b,c, breaking ties in order a,b,c.
	if pa <= pb && pa <= pc {
		return a
	} else if pb <= pc {
		return b
	}
	return c
}

func colorDifference(a, b color.Color) colorDelta {
	var ca, cb [4]uint32
	ca[0], ca[1], ca[2], ca[3] = a.RGBA()
	cb[0], cb[1], cb[2], cb[3] = b.RGBA()

	const full = 65535
	var delta [4]int32
	for i := 0; i < 3; i++ {
		pa := ca[i] * full
		if ca[3] > 0 {
			pa /= ca[3]
		}
		pb := cb[i] * full
		if cb[3] > 0 {
			pb /= cb[3]
		}
		delta[i] = int32(pa) - int32(pb)
	}
	delta[3] = int32(ca[3]) - int32(cb[3])

	/*
	 * Compute a very basic perceptual distance using
	 * formula from http://www.compuphase.com/cmetric.htm .
	 */
	redA := ca[0]
	redB := cb[0]
	if ca[3] > 0 {
		redA = redA * full / ca[3]
	}
	if cb[3] > 0 {
		redB = redB * full / cb[3]
	}

	redMean := int32((redA + redB) / 2)
	return colorDelta{
		int32((2*full + redMean) * delta[0] / (3 * full)),
		int32(4 * delta[1] / 3),
		int32((3*full - redMean) * delta[2] / (3 * full)),
		int32(delta[3]),
	}
}

func (a colorDelta) magnitude() uint64 {
	var d2 uint64
	for i := 0; i < deltaComponents; i++ {
		d2 += uint64(int64(a[i]) * int64(a[i]))
	}

	return d2
}

func (a colorDelta) add(b colorDelta) colorDelta {
	var delta colorDelta
	for i := 0; i < deltaComponents; i++ {
		delta[i] = a[i] + b[i]
	}
	return delta
}

func diffuseColorDeltas(colorError [3][]colorDelta, x int) colorDelta {
	var delta colorDelta
	// Sierra dithering
	for i := 0; i < deltaComponents; i++ {
		delta[i] += 2 * colorError[2][x-1][i]
		delta[i] += 3 * colorError[2][x][i]
		delta[i] += 2 * colorError[2][x+1][i]
		delta[i] += 2 * colorError[1][x-2][i]
		delta[i] += 4 * colorError[1][x-1][i]
		delta[i] += 5 * colorError[1][x][i]
		delta[i] += 4 * colorError[1][x+1][i]
		delta[i] += 2 * colorError[1][x+2][i]
		delta[i] += 3 * colorError[0][x-2][i]
		delta[i] += 5 * colorError[0][x-1][i]
		if delta[i] < 0 {
			delta[i] -= 16
		} else {
			delta[i] += 16
		}
		delta[i] /= 32
	}
	return delta
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
