package pixel

import (
	"image/color"
	"math"
)

type Pixel struct {
	// RGB comes from Color
	R uint32
	G uint32
	B uint32
	A uint32
	X int
	Y int
	// HSV and lum are calculated
	h    float64
	s    float64
	v    float64
	h2   int64
	lum2 int64
	v2   int64
}

type AGreaterThanB struct {
	Name string
	Exec func(a, b *Pixel) bool
}

func (p *Pixel) Init(c color.Color, x, y int) *Pixel {
	R, G, B, _ := c.RGBA()
	p.R = R
	p.G = G
	p.B = B
	p.X = x
	p.Y = y
	p.A = math.MaxUint32

	p.calcHSV()
	p.calcLum()
	return p
}

func (p Pixel) RGBA() (r, g, b, a uint32) {
	r = p.R
	g = p.G
	b = p.B
	a = math.MaxUint32
	return
}

func (p *Pixel) calcHSV() {
	rprime := float64(p.R) / 255.0
	gprime := float64(p.G) / 255.0
	bprime := float64(p.B) / 255.0
	cmax := math.Max(rprime, math.Max(gprime, bprime))
	cmin := math.Min(rprime, math.Min(gprime, bprime))
	delta := cmax - cmin

	var h, s, v float64

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

	p.h = h
	p.s = s
	p.v = v
}

func (p *Pixel) calcLum() {
	lum := math.Sqrt(0.241*float64(p.R) + .691*float64(p.G) + 0.068*float64(p.B))
	p.h2 = int64(p.h * 8)
	p.lum2 = int64(lum * 8)
	p.v2 = int64(p.v * 8)

	if p.h2%2 == 1 {
		p.v2 = 8 - p.v2
		p.lum2 = 8 - p.lum2
	}
}

func StepHSVGreaterThan(a, b *Pixel) bool {

	if a.h > b.h {
		return true
	} else if a.h2 < b.h2 {
		return false
	} else if a.lum2 > b.lum2 {
		return true
	} else if a.lum2 < b.lum2 {
		return false
	} else if a.v2 > b.v2 {
		return true
	}
	return false
}

func HSVGreaterThan(a, b *Pixel) bool {
	if a.h > b.h {
		return true
	} else if a.h < b.h {
		return false
	} else if a.s > b.s {
		return true
	} else if a.s < b.s {
		return false
	} else if a.v > b.v {
		return true
	}
	return false
}

func GreaterThan(a, b *Pixel) bool {
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
