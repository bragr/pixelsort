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

type pixel struct {
	R    uint32
	G    uint32
	B    uint32
	H    float64
	S    float64
	V    float64
	H2   int64
	Lum2 int64
	V2   int64
}

func makePixel(c color.Color) pixel {
	R, G, B, _ := c.RGBA()
	H, S, V := rgbToHSV(R, G, B)
	H2, Lum2, V2 := luminosity(R, G, B, H, V)
	return pixel{R, G, B, H, S, V, H2, Lum2, V2}
}

func (p pixel) RGBA() (r, g, b, a uint32) {
	r = p.R
	g = p.G
	b = p.B
	a = math.MaxUint32
	return
}

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

func rgbToHSV(r, g, b uint32) (h, s, v float64) {
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

func luminosity(r, g, b uint32, h, v float64) (h2, lum2, v2 int64) {
	lum := math.Sqrt(0.241*float64(r) + .691*float64(g) + 0.068*float64(b))
	h2 = int64(h * 8)
	lum2 = int64(lum * 8)
	v2 = int64(v * 8)

	if h2%2 == 1 {
		v2 = 8 - v2
		lum2 = 8 - lum2
	}
	return
}

func StepHSVaGreaterThanB(a, b pixel) bool {
	if a.H > b.H {
		return true
	} else if a.H2 < b.H2 {
		return false
	} else if a.Lum2 > b.Lum2 {
		return true
	} else if a.Lum2 < b.Lum2 {
		return false
	} else if a.V2 > b.V2 {
		return true
	}
	return false
}

func HSVaGreaterThanB(a, b pixel) bool {
	if a.H > b.H {
		return true
	} else if a.H < b.H {
		return false
	} else if a.S > b.S {
		return true
	} else if a.S < b.S {
		return false
	} else if a.V > b.V {
		return true
	}
	return false
}

func aGreaterThanB(a, b color.Color) bool {
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	if ar > br {
		return true
	} else if ar < br {
		return false
	} else if ag > bg {
		return true
	} else if ag < bg {
		return false
	} else if ab > bb {
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

func toImage(p []pixel, out draw.Image) draw.Image {
	maxX := out.Bounds().Max.X
	maxY := out.Bounds().Max.Y
	for i := 0; i < len(p); i++ {
		x, y := toXY(i, maxX, maxY)
		out.Set(x, y, p[i])
	}
	return out
}

func selectionSortI(in image.Image) image.Image {
	step := 0
	out := in.(draw.Image)

	// Copy data
	data := make([]pixel, in.Bounds().Max.X*in.Bounds().Max.Y)
	for i := 0; i < len(data); i++ {
		x, y := toXY(i, in.Bounds().Max.X, in.Bounds().Max.Y)
		data[i] = makePixel(in.At(x, y))
	}

	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	for slot := len(data) - 1; slot >= 0; slot-- {
		max := 0
		for i := 1; i <= slot; i++ {
			if StepHSVaGreaterThanB(data[i], data[max]) {
				max = i
			}
		}
		if max != slot {
			tmp := data[slot]
			data[slot] = data[max]
			data[max] = tmp
		}
		fmt.Printf("Step %06d\n", step)
		if step%100 == 0 {
			filename := fmt.Sprintf("s%09d.png", step)
			outfile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0665)
			handleError(err)
			err = enc.Encode(outfile, toImage(data, out))
			handleError(err)
		}
		step++
	}
	return out
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
