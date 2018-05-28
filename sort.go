package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"
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

type PixelAGreaterThanB func(a, b pixel) bool

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

func AGreaterThanB(a, b pixel) bool {
	if a.R > b.R {
		return true
	} else if a.R < b.R {
		return false
	} else if a.G > b.G {
		return true
	} else if a.G < b.G {
		return false
	} else if a.B > b.B {
		return true
	}
	return false
}

func toXY(i, xSize, ySize int) (x, y int) {
	y = i % ySize
	x = i / ySize
	return x, y
}

func toImage(p *[]pixel, out *draw.Image) *draw.Image {
	maxX := (*out).Bounds().Max.X
	maxY := (*out).Bounds().Max.Y
	for i := 0; i < len(*p); i++ {
		x, y := toXY(i, maxX, maxY)
		(*out).Set(x, y, (*p)[i])
	}
	return out
}

func copyImage(in image.Image) *[]pixel {
	// Copy data
	data := make([]pixel, in.Bounds().Max.X*in.Bounds().Max.Y)
	for i := 0; i < len(data); i++ {
		x, y := toXY(i, in.Bounds().Max.X, in.Bounds().Max.Y)
		data[i] = makePixel(in.At(x, y))
	}
	return &data
}

func writeStep(prefix string, step int, stepLimit *int, enc *png.Encoder, out *draw.Image, data *[]pixel) {
	if step%*stepLimit == 0 {
		fmt.Printf("Step %06d\n", step)
		filename := fmt.Sprintf("%s_%09d.png", prefix, step)
		outfile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0665)
		handleError(err)
		err = enc.Encode(outfile, *toImage(data, out))
		handleError(err)
	}
}

func insertionSort(in image.Image, compares *[]PixelAGreaterThanB, stepLimit *int) {
	out := in.(draw.Image)
	enc := &png.Encoder{CompressionLevel: png.BestSpeed}

	for index, compare := range *compares {
		step := 0
		data := *copyImage(in)
		fmt.Printf("Print with compare #%d\n", index)
		prefix := fmt.Sprint("insertion_%d", index)

		for i := 0; i < len(data); i++ {
			for j := i; j > 0 && compare(data[j-1], data[j]); j-- {
				tmp := data[j-1]
				data[j-1] = data[j]
				data[j] = tmp
				writeStep(prefix, step, stepLimit, enc, &out, &data)
				step++
			}
		}
	}
}

func selectionSort(in image.Image, compares *[]PixelAGreaterThanB, stepLimit *int) {
	out := in.(draw.Image)
	enc := &png.Encoder{CompressionLevel: png.BestSpeed}

	for index, compare := range *compares {
		step := 0
		data := *copyImage(in)
		fmt.Printf("Print with compare #%d\n", index)
		prefix := fmt.Sprint("selection_%d", index)

		for slot := len(data) - 1; slot >= 0; slot-- {
			max := 0
			for i := 1; i <= slot; i++ {
				if compare(data[i], data[max]) {
					max = i
				}
			}
			if max != slot {
				tmp := data[slot]
				data[slot] = data[max]
				data[max] = tmp
			}
			writeStep(prefix, step, stepLimit, enc, &out, &data)
			step++
		}
	}
}

func main() {
	fmt.Println("Hello sort!")
	defer fmt.Println("Goodbye sort!")

	//Basic Options
	filename := flag.String("f", "", "Input file")
	frameStep := flag.Int("step", 100, "How often to output a frame")

	//Which sorts to run
	enableInsertion := flag.Bool("insertion", false, "Enable insertion sort")
	enableSelection := flag.Bool("selection", false, "Enable selection sort")

	//Which comparisons to use
	enableStepHSV := flag.Bool("stephsv", false, "Enable STEP HSV comparison")
	enableHSV := flag.Bool("hsv", false, "Enable HSV comparison")
	enableSimple := flag.Bool("simple", false, "Enable simple comparison")

	flag.Parse()

	// Open the Image
	inFile, err := os.Open(*filename)
	handleError(err)
	origImage, _, err := image.Decode(inFile)
	handleError(err)

	compare := []PixelAGreaterThanB{}
	if *enableStepHSV {
		compare = append(compare, StepHSVaGreaterThanB)
	}
	if *enableHSV {
		compare = append(compare, HSVaGreaterThanB)
	}
	if *enableSimple {
		compare = append(compare, AGreaterThanB)
	}

	if *enableInsertion {
		insertionSort(origImage, &compare, frameStep)
	}
	if *enableSelection {
		selectionSort(origImage, &compare, frameStep)
	}
}
