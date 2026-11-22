// Package art draws every image on the site in code: the bedroom in each
// lighting phase, title art, the 88x31 button, the cursor and the favicon.
// Nothing is traced or copied; `go run ./cmd/artgen` writes the PNGs and CI
// checks they match this code (ADR-004).
package art

import (
	"image"
	"image/color"
)

// RGB is an opaque color.
type RGB struct{ R, G, B uint8 }

// Hex parses "#rrggbb". It panics on bad input because every color is a
// constant in this package.
func Hex(s string) RGB {
	if len(s) != 7 || s[0] != '#' {
		panic("art: bad color " + s)
	}
	v := func(i int) uint8 {
		h := func(c byte) uint8 {
			switch {
			case c >= '0' && c <= '9':
				return c - '0'
			case c >= 'a' && c <= 'f':
				return c - 'a' + 10
			case c >= 'A' && c <= 'F':
				return c - 'A' + 10
			}
			panic("art: bad color " + s)
		}
		return h(s[i])<<4 | h(s[i+1])
	}
	return RGB{v(1), v(3), v(5)}
}

// Layer is a grid of optional pixels; ok marks which ones are set.
type Layer struct {
	W, H int
	pix  []RGB
	ok   []bool
}

// NewLayer returns an empty (fully transparent) layer.
func NewLayer(w, h int) *Layer {
	return &Layer{W: w, H: h, pix: make([]RGB, w*h), ok: make([]bool, w*h)}
}

// Set paints one pixel; out-of-bounds writes are ignored.
func (l *Layer) Set(x, y int, c RGB) {
	if x < 0 || y < 0 || x >= l.W || y >= l.H {
		return
	}
	l.pix[y*l.W+x] = c
	l.ok[y*l.W+x] = true
}

// At returns a pixel and whether it is set.
func (l *Layer) At(x, y int) (RGB, bool) {
	if x < 0 || y < 0 || x >= l.W || y >= l.H {
		return RGB{}, false
	}
	return l.pix[y*l.W+x], l.ok[y*l.W+x]
}

// Rect fills a rectangle.
func (l *Layer) Rect(x, y, w, h int, c RGB) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			l.Set(i, j, c)
		}
	}
}

// Clear unsets a rectangle.
func (l *Layer) Clear(x, y, w, h int) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			if i >= 0 && j >= 0 && i < l.W && j < l.H {
				l.ok[j*l.W+i] = false
			}
		}
	}
}

// Frame draws a 1px rectangle outline.
func (l *Layer) Frame(x, y, w, h int, c RGB) {
	l.Rect(x, y, w, 1, c)
	l.Rect(x, y+h-1, w, 1, c)
	l.Rect(x, y, 1, h, c)
	l.Rect(x+w-1, y, 1, h, c)
}

// Checker fills a rectangle with a two-color dither.
func (l *Layer) Checker(x, y, w, h int, a, b RGB) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			if (i+j)%2 == 0 {
				l.Set(i, j, a)
			} else {
				l.Set(i, j, b)
			}
		}
	}
}

// Line draws a 1px line (Bresenham).
func (l *Layer) Line(x0, y0, x1, y1 int, c RGB) {
	dx, sx := abs(x1-x0), 1
	if x0 > x1 {
		sx = -1
	}
	dy, sy := -abs(y1-y0), 1
	if y0 > y1 {
		sy = -1
	}
	e := dx + dy
	for {
		l.Set(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * e
		if e2 >= dy {
			e += dy
			x0 += sx
		}
		if e2 <= dx {
			e += dx
			y0 += sy
		}
	}
}

// Ellipse fills an axis-aligned ellipse centered on (cx, cy).
func (l *Layer) Ellipse(cx, cy, rx, ry int, c RGB) {
	for j := -ry; j <= ry; j++ {
		for i := -rx; i <= rx; i++ {
			if float64(i*i)/float64(rx*rx+1)+float64(j*j)/float64(ry*ry+1) <= 1 {
				l.Set(cx+i, cy+j, c)
			}
		}
	}
}

// Sprite draws rows of runes; each rune maps to a color in pal, and runes
// missing from pal (use '.') are transparent.
func (l *Layer) Sprite(x, y int, rows []string, pal map[rune]RGB) {
	for j, row := range rows {
		for i, r := range []rune(row) {
			if c, ok := pal[r]; ok {
				l.Set(x+i, y+j, c)
			}
		}
	}
}

// Draw copies every set pixel of src onto l at (x, y).
func (l *Layer) Draw(src *Layer, x, y int) {
	for j := 0; j < src.H; j++ {
		for i := 0; i < src.W; i++ {
			if c, ok := src.At(i, j); ok {
				l.Set(x+i, y+j, c)
			}
		}
	}
}

// Image converts the layer to an RGBA image; unset pixels are transparent.
func (l *Layer) Image() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, l.W, l.H))
	for j := 0; j < l.H; j++ {
		for i := 0; i < l.W; i++ {
			if c, ok := l.At(i, j); ok {
				img.SetRGBA(i, j, color.RGBA{c.R, c.G, c.B, 255})
			}
		}
	}
	return img
}

// Scale returns a copy enlarged by an integer factor (nearest neighbor).
func (l *Layer) Scale(k int) *Layer {
	out := NewLayer(l.W*k, l.H*k)
	for j := 0; j < l.H; j++ {
		for i := 0; i < l.W; i++ {
			if c, ok := l.At(i, j); ok {
				out.Rect(i*k, j*k, k, k, c)
			}
		}
	}
	return out
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func clamp8(v float64) uint8 {
	switch {
	case v < 0:
		return 0
	case v > 255:
		return 255
	}
	return uint8(v + 0.5)
}

// mix blends a toward b by t in [0, 1].
func mix(a, b RGB, t float64) RGB {
	return RGB{
		clamp8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		clamp8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		clamp8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
	}
}
