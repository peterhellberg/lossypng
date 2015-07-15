package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"path"
	"strings"

	_ "image/gif" // for image.Decode() format registration
	_ "image/jpeg"
	"image/png"

	"github.com/peterhellberg/lossypng"
)

func main() {
	var (
		convertToRGBA      bool
		convertToGrayscale bool
		quantization       int
		extension          string
	)

	flag.BoolVar(&convertToRGBA, "c", false, "convert image to 32-bit color")
	flag.BoolVar(&convertToGrayscale, "g", false, "convert image to grayscale")
	flag.IntVar(&quantization, "s", 10, "quantization threshold, zero is lossless")
	flag.StringVar(&extension, "e", "-lossy.png", "filename extension of output files")

	flag.Parse()

	var colorConversion int
	if convertToRGBA && !convertToGrayscale {
		colorConversion = lossypng.RGBAConversion
	} else if convertToGrayscale && !convertToRGBA {
		colorConversion = lossypng.GrayscaleConversion
	}

	if args := flag.Args(); len(args) > 0 {
		fn := args[0]

		f, err := os.Open(fn)
		if err != nil {
			fmt.Printf("couldn't open %v: %v\n", fn, err)
			return
		}

		inInfo, err := f.Stat()
		if err != nil {
			fmt.Printf("couldn't stat %v: %v\n", fn, err)
			return
		}

		m, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			fmt.Printf("couldn't decode %v: %v\n", fn, err)
			return
		}

		o := lossypng.Optimize(m, colorConversion, quantization)

		// save optimized image
		outPath := pathWithSuffix(fn, extension)
		outFile, createErr := os.Create(outPath)
		if createErr != nil {
			fmt.Printf("couldn't create %v: %v\n", outPath, createErr)
			return
		}

		if err := png.Encode(outFile, o); err != nil {
			fmt.Printf("couldn't encode %v: %v\n", fn, err)
			return
		}

		outInfo, err := outFile.Stat()
		if err != nil {
			fmt.Printf("couldn't stat %v: %v\n", outFile, err)
			return
		}
		outFile.Close()

		var (
			inSize      = inInfo.Size()
			inSizeDesc  = sizeDesc(inSize)
			outSize     = outInfo.Size()
			outSizeDesc = sizeDesc(outSize)
			percentage  = fmt.Sprintf("%d%%", (outSize*100+inSize/2)/inSize)
		)

		fmt.Printf("compressed %s (%s) to %s (%s, %s)\n",
			path.Base(fn), inSizeDesc, path.Base(outPath), outSizeDesc, percentage)
	}
}

func pathWithSuffix(filePath string, suffix string) string {
	extension := path.Ext(filePath)
	insertion := len(extension)
	if insertion > 0 {
		// if extension exists, trim it off of the base filename
		insertion = strings.LastIndex(filePath, extension)
	} else {
		insertion = len(filePath)
	}
	return filePath[:insertion] + suffix
}

func sizeDesc(size int64) string {
	suffixes := []string{"B", "kB", "MB", "GB", "TB"}
	var i int
	for i = 0; i+1 < len(suffixes); i++ {
		if size < 10000 {
			break
		}
		size = (size + 500) / 1000
	}
	return fmt.Sprintf("%d%v", size, suffixes[i])
}