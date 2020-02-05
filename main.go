package main

import (
	"github.com/kettek/apng"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"os"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()

	g, err := gif.DecodeAll(f)
	if err != nil {
		panic(err)
	}

	// fix gif compression technique that uses b frames
	for i := 1; i < len(g.Image); i++ {
		src := g.Image[i-1]
		lyr := g.Image[i]
		b := src.Bounds()
		dst := image.NewPaletted(b, src.Palette)
		draw.Draw(dst, b, src, b.Min, draw.Src)
		draw.Draw(dst, b, lyr, b.Min, draw.Over)

		g.Image[i] = dst
	}

	a := apng.APNG{
		Frames: make([]apng.Frame, len(g.Image)),
	}
	a.LoopCount = uint(g.LoopCount)

	for i := range g.Image {
		img := g.Image[i]
		b := img.Bounds()
		nImg := image.NewRGBA(b)

		for x := 0; x < b.Max.X; x++ {
			for y := 0; y < b.Max.Y; y++ {
				oldPixel := img.At(x, y)
				r, g, b, _ := oldPixel.RGBA()
				lum := uint8((19595*r + 38470*g + 7471*b + 1<<15) >> 24)
				nImg.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: math.MaxUint8 - lum})
			}
		}

		f := apng.Frame{
			Image:            nImg,
			DelayNumerator:   uint16(g.Delay[i]),
			DelayDenominator: 100,
		}
		switch g.Disposal[i] {
		case gif.DisposalNone:
			f.DisposeOp = apng.DISPOSE_OP_NONE
		case gif.DisposalBackground:
			f.DisposeOp = apng.DISPOSE_OP_BACKGROUND
		case gif.DisposalPrevious:
			f.DisposeOp = apng.DISPOSE_OP_PREVIOUS
		}
		a.Frames[i] = f
	}

	out, err := os.Create("output.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	apng.Encode(out, a)
}
