package main

import (
	"./pixel"
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
)

func handleError(e error) {
	if e != nil {
		panic(e)
	}
}

func toXY(i, xSize, ySize int) (x, y int) {
	y = i % ySize
	x = i / ySize
	return x, y
}

func copyImage(in image.Image) *[]pixel.Pixel {
	// Copy data
	data := make([]pixel.Pixel, in.Bounds().Max.X*in.Bounds().Max.Y)
	for i := 0; i < len(data); i++ {
		x, y := toXY(i, in.Bounds().Max.X, in.Bounds().Max.Y)
		data[i] = *(new(pixel.Pixel).Init(in.At(x, y), x, y))
	}
	return &data
}

func writeStep(prefix string, step int, stepLimit *int, enc *png.Encoder, out *draw.Image) {
	if step%*stepLimit == 0 {
		fmt.Printf("Step %06d\n", step)

		filename := fmt.Sprintf("%s_%09d.png", prefix, step)
		outfile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0665)
		handleError(err)

		err = enc.Encode(outfile, *out)
		handleError(err)
	}
}

func swap(i, j, maxX, maxY int, data *[]pixel.Pixel, out *draw.Image) {
	tmp := (*data)[j]

	(*data)[j] = (*data)[i]
	x, y := toXY(j, maxX, maxY)
	(*out).Set(x, y, (*data)[i])

	(*data)[i] = tmp
	x, y = toXY(i, maxX, maxY)
	(*out).Set(x, y, tmp)
}

func insertionSort(in image.Image, compares *[]pixel.AGreaterThanB, stepLimit *int) {
	maxX := in.Bounds().Max.X
	maxY := in.Bounds().Max.Y
	enc := &png.Encoder{CompressionLevel: png.BestSpeed}

	for _, compare := range *compares {
		step := 0
		data := *copyImage(in)
		out := in.(draw.Image)
		fmt.Printf("Print with compare %s\n", compare.Name)
		prefix := fmt.Sprintf("insertion_%s", compare.Name)

		for i := 0; i < len(data); i++ {
			for j := i; j > 0 && compare.Exec(&data[j-1], &data[j]); j-- {
				swap(j, j-1, maxX, maxY, &data, &out)
				writeStep(prefix, step, stepLimit, enc, &out)
				step++
			}
		}
	}
}

func selectionSort(in image.Image, compares *[]pixel.AGreaterThanB, stepLimit *int) {
	maxX := in.Bounds().Max.X
	maxY := in.Bounds().Max.Y
	enc := &png.Encoder{CompressionLevel: png.BestSpeed}

	for _, compare := range *compares {
		step := 0
		data := *copyImage(in)
		out := in.(draw.Image)
		fmt.Printf("Print with compare %s\n", compare.Name)
		prefix := fmt.Sprintf("selection_%s", compare.Name)

		for slot := len(data) - 1; slot >= 0; slot-- {
			max := 0
			for i := 1; i <= slot; i++ {
				if compare.Exec(&data[i], &data[max]) {
					max = i
				}
			}
			if max != slot {
				swap(max, slot, maxX, maxY, &data, &out)
			}
			writeStep(prefix, step, stepLimit, enc, &out)
			step++
		}
	}
}

func BubbleSort(in image.Image, compares *[]pixel.AGreaterThanB, stepLimit *int) {
	maxX := in.Bounds().Max.X
	maxY := in.Bounds().Max.Y
	enc := &png.Encoder{CompressionLevel: png.BestSpeed}

	for _, compare := range *compares {
		step := 0
		data := *copyImage(in)
		out := in.(draw.Image)
		fmt.Printf("Print with compare %s\n", compare.Name)
		prefix := fmt.Sprintf("bubble_%s", compare.Name)

		next := 0
		for n := len(data); n > 0; {
			next = 0
			for i := 1; i < n; i++ {
				if compare.Exec(&data[i-1], &data[i]) {
					swap(i, i-1, maxX, maxY, &data, &out)
					next = i
				}
			}
			n = next
			step++
			writeStep(prefix, step, stepLimit, enc, &out)
		}
	}
}

func splitAndMerge(b, a *[]pixel.Pixel, begin, end int, compare pixel.AGreaterThanB, out *draw.Image, step *int, stepLimit *int, maxX int, maxY int, prefix *string, enc *png.Encoder) {
	if end-begin < 2 {
		return
	}

	// split
	middle := (end + begin) / 2
	splitAndMerge(a, b, begin, middle, compare, out, step, stepLimit, maxX, maxY, prefix, enc)
	splitAndMerge(a, b, middle, end, compare, out, step, stepLimit, maxX, maxY, prefix, enc)

	(*step)++

	// merge
	i := begin
	j := middle
	for k := begin; k < end; k++ {
		if i < middle && (j >= end || compare.Exec(&(*b)[j], &(*b)[i])) {
			(*a)[k] = (*b)[i]
			x, y := toXY(k, maxX, maxY)
			(*out).Set(x, y, (*b)[i])
			i = i + 1
		} else {
			(*a)[k] = (*b)[j]
			x, y := toXY(k, maxX, maxY)
			(*out).Set(x, y, (*b)[j])
			j = j + 1
		}
	}
	writeStep(*prefix, *step, stepLimit, enc, out)
}

func MergeSort(in image.Image, compares *[]pixel.AGreaterThanB, stepLimit *int) {
	maxX := in.Bounds().Max.X
	maxY := in.Bounds().Max.Y
	enc := &png.Encoder{CompressionLevel: png.BestSpeed}

	for _, compare := range *compares {
		step := 0
		data := copyImage(in)
		scratch := copyImage(in)
		out := in.(draw.Image)
		fmt.Printf("Print with compare %s\n", compare.Name)
		prefix := fmt.Sprintf("merge_%s", compare.Name)

		splitAndMerge(data, scratch, 0, len(*data), compare, &out, &step, stepLimit, maxX, maxY, &prefix, enc)
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
	enableBubble := flag.Bool("bubble", false, "Enable bubble sort")
	enableMerge := flag.Bool("merge", false, "Enable merge sort")

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

	compare := []pixel.AGreaterThanB{}
	if *enableStepHSV {
		compare = append(compare, pixel.AGreaterThanB{"step_hsv", pixel.StepHSVGreaterThan})
	}
	if *enableHSV {
		compare = append(compare, pixel.AGreaterThanB{"hsv", pixel.HSVGreaterThan})
	}
	if *enableSimple {
		compare = append(compare, pixel.AGreaterThanB{"simple", pixel.GreaterThan})
	}

	if *enableInsertion {
		insertionSort(origImage, &compare, frameStep)
	}
	if *enableSelection {
		selectionSort(origImage, &compare, frameStep)
	}
	if *enableBubble {
		BubbleSort(origImage, &compare, frameStep)
	}
	if *enableMerge {
		MergeSort(origImage, &compare, frameStep)
	}
}
