package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"math/rand"
	"os"
)

func handleError(e error) {
	if e != nil {
		panic(e)
	}
}

func seedRandom() {
	f, err := os.Open("/dev/urandom")
	handleError(err)
	reader := bufio.NewReader(f)

	seedArray := make([]byte, 8)
	_, err = io.ReadFull(reader, seedArray)
	handleError(err)
	seed := binary.LittleEndian.Uint64(seedArray)
	rand.Seed(int64(seed))
}

func rgbToHSV(c color.Color) (h, s, v float64) {
	r, g, b, _ := c.RGBA()
	rprime := float64(r) / 255.0
	gprime := float64(g) / 255.0
	bprime := float64(b) / 255.0
	cmax := math.Max(rprime, math.Max(gprime, bprime))
	cmin := math.Min(rprime, math.Min(gprime, bprime))
	delta := cmax - cmin

	if delta == 0 {
		h = 0.0
	} else if cmax == rprime {
		h = 60 * math.Mod((gprime-bprime)/delta, 6.0)
	} else if cmax == gprime {
		h = 60 * ((bprime-rprime)/delta + 2)
	} else {
		h = 60 * ((rprime-gprime)/delta + 4)
	}
	if cmax == 0 {
		s = 0.0
	} else {
		s = delta / cmax
	}
	v = cmax

	return h, s, v
}

func aGreaterThanB(a, b color.Color) bool {
	ah, as, av := rgbToHSV(a)
	bh, bs, bv := rgbToHSV(b)
	if ah > bh {
		return true
	} else if ah < bh {
		return false
	} else if as > bs {
		return true
	} else if as < bs {
		return false
	} else if av > bv {
		return true
	}
	return false
}

func toXY(i, xSize, ySize int) (x, y int) {
	y = i % ySize
	x = i / ySize
	return x, y
}

func insertionSort(data []byte) []byte {
	sorted := make([]byte, len(data))
	copy(sorted, data)
	step := 0
	for i := 0; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j-1] > sorted[j]; j-- {
			tmp := sorted[j-1]
			sorted[j-1] = sorted[j]
			sorted[j] = tmp
			fmt.Printf("Step %04d: %v\n", step, sorted)
			step++
		}
	}
	return sorted
}

func selectionSortI(in image.Image) image.Image {
	step := 0
	data := in.(draw.Image)
	iMaxX := data.Bounds().Max.X
	iMaxY := data.Bounds().Max.Y
	fmt.Printf("X: %d\t Y: %d\n", iMaxX, iMaxY)
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	for slot := iMaxX*iMaxY - 1; slot >= 0; slot-- {
		max := 0
		for i := 1; i <= slot; i++ {
			maxX, maxY := toXY(max, iMaxX, iMaxY)
			iX, iY := toXY(i, iMaxX, iMaxY)
			if aGreaterThanB(data.At(iX, iY), data.At(maxX, maxY)) {
				max = i
			}
		}
		if max != slot {
			slotX, slotY := toXY(slot, iMaxX, iMaxY)
			maxX, maxY := toXY(max, iMaxX, iMaxY)
			tmp := data.At(slotX, slotY)
			data.Set(slotX, slotY, data.At(maxX, maxY))
			data.Set(maxX, maxY, tmp)
		}
		fmt.Printf("Step %06d\n", step)
		if step%100 == 0 {
			filename := fmt.Sprintf("%09d.png", step)
			out, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0665)
			handleError(err)
			err = enc.Encode(out, data)
			handleError(err)
		}
		step++
	}
	return data
}

func main() {
	fmt.Println("Hello sort!")
	defer fmt.Println("Goodbye sort!")
	seedRandom()

	if len(os.Args) < 2 {
		fmt.Println("FEED ME MORE ARGS")
		os.Exit(1)
	}
	filename := os.Args[1]
	inFile, err := os.Open(filename)
	handleError(err)
	origImage, _, err := image.Decode(inFile)
	handleError(err)

	maxX := origImage.Bounds().Max.X
	maxY := origImage.Bounds().Max.Y
	fmt.Printf("X: %d\t Y: %d\n", maxX, maxY)

	_ = selectionSortI(origImage)

}
