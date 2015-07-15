***Warning: This library should not be used just yet***

# lossypng

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/peterhellberg/lossypng)

Library version of the [lossypng](https://github.com/foobaz/lossypng) command line tool.

> Shrink PNG files by applying a lossy filter

## Installation

    go get -u github.com/peterhellberg/lossypng

Feel free to copy all or parts of this package into your own codebase.

## Examples

### Original image `320 KB`

![Original image](http://assets.c7.se/lossypng/dakar-original.png)

The optimized images were encoded to PNG using a `png.Encoder{png.BestCompression}`

### Optimize(m, RGBAConversion, 10) `156 KB`

![Optimize(m, RGBAConversion, 10)](http://assets.c7.se/lossypng/dakar-rgba-10.png)

### Optimize(m, GrayscaleConversion, 10) `40 KB`

![Optimize(m, GrayscaleConversion, 10)](http://assets.c7.se/lossypng/dakar-grayscale-10.png)

## Credit

This compression technique was invented by Michael Vinther for his excellent
Windows program, [Image Analyzer](http://meesoft.logicnet.dk/Analyzer/). It
does much more than just compression. It was ported and improved by
[William MacKay](https://github.com/foobaz/lossypng).

I have just converted it into a library.

## License

All code in lossypng is public domain. You may use it however you wish.
