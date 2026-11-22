package art

import (
	"fmt"
	"strings"
)

// Title is one page heading drawn as WordArt-style title art.
type Title struct {
	File string // PNG name without extension
	Text string
	Top  RGB // gradient top
	Bot  RGB // gradient bottom
}

// Titles for every page. Scale 6 for the site banner, 4 for page titles.
var Titles = []Title{
	{"title-home", "GAMER'S PARADISE", Hex("#fff27a"), Hex("#ff5a1e")},
	{"title-xbox", "THE XBOX YEARS", Hex("#c8ff7a"), Hex("#2aa02a")},
	{"title-gameboy", "GAME BOY", Hex("#e0f8a0"), Hex("#4f9a2c")},
	{"title-shelf", "THE SHELF", Hex("#ffe0a0"), Hex("#b86a2a")},
	{"title-now", "RIGHT NOW", Hex("#ffb0d8"), Hex("#8a2ac8")},
	{"title-box", "COMING SOON", Hex("#fff27a"), Hex("#e8a020")},
	{"title-guestbook", "GUESTBOOK", Hex("#b0e0ff"), Hex("#2a5ad8")},
	{"title-arcade", "ARCADE", Hex("#fff27a"), Hex("#e83a8a")},
	{"title-skate", "HALF-PIPE HERO", Hex("#a0f0ff"), Hex("#e8803a")},
	{"title-links", "TOP 8", Hex("#d0c0ff"), Hex("#5a3ab8")},
	{"title-404", "404", Hex("#ffb0a0"), Hex("#c82a2a")},
}

// TitleArt renders text at an integer scale with a vertical gradient, a
// dark outline and a drop shadow.
func TitleArt(t Title, scale int) *Layer {
	mask := NewLayer(TextWidth(t.Text), 7)
	mask.Text(0, 0, t.Text, cWhite)
	big := mask.Scale(scale)
	pad := scale
	out := NewLayer(big.W+pad*2+scale, big.H+pad*2+scale)
	outline := Hex("#1a0a2a")
	shadow := Hex("#3a1a4a")
	// drop shadow
	for y := 0; y < big.H; y++ {
		for x := 0; x < big.W; x++ {
			if _, ok := big.At(x, y); ok {
				out.Set(x+pad+scale, y+pad+scale, shadow)
			}
		}
	}
	// outline: any empty pixel next to a filled one
	o := scale / 2
	if o < 1 {
		o = 1
	}
	for y := -o; y < big.H+o; y++ {
		for x := -o; x < big.W+o; x++ {
			if _, ok := big.At(x, y); ok {
				continue
			}
			near := false
			for dy := -o; dy <= o && !near; dy++ {
				for dx := -o; dx <= o; dx++ {
					if _, ok := big.At(x+dx, y+dy); ok {
						near = true
						break
					}
				}
			}
			if near {
				out.Set(x+pad, y+pad, outline)
			}
		}
	}
	for y := 0; y < big.H; y++ {
		c := mix(t.Top, t.Bot, float64(y)/float64(big.H-1))
		if y%scale == 0 && y/scale == 1 {
			c = mix(c, cWhite, 0.5) // shine line
		}
		for x := 0; x < big.W; x++ {
			if _, ok := big.At(x, y); ok {
				out.Set(x+pad, y+pad, c)
			}
		}
	}
	return out
}

// Button is the site's 88x31 badge for other people's link pages.
func Button() *Layer {
	l := NewLayer(88, 31)
	l.Rect(0, 0, 88, 31, Hex("#1a1440"))
	for y := 0; y < 31; y++ {
		l.Rect(0, y, 88, 1, mix(Hex("#2a2070"), Hex("#0a0820"), float64(y)/30))
	}
	l.Frame(0, 0, 88, 31, Hex("#8a7ae8"))
	l.Rect(1, 1, 86, 1, Hex("#c8c0ff"))
	l.Rect(1, 1, 1, 29, Hex("#c8c0ff"))
	// little CRT
	l.Rect(4, 6, 22, 18, cGrey)
	l.Rect(6, 8, 18, 13, Hex("#6ab0f0"))
	l.Rect(6, 17, 18, 4, Hex("#d8a868"))
	l.Rect(12, 13, 2, 4, cOrange)
	l.Rect(10, 24, 10, 2, cGreyDark)
	l.Text(30, 5, "GAMER'S", Hex("#fff27a"))
	l.Text(30, 17, "PARADISE", Hex("#ff9a4a"))
	return l
}

// Cursor is a classic pixel arrow (12x19).
func Cursor() *Layer {
	l := NewLayer(12, 19)
	l.Sprite(0, 0, []string{
		"k...........",
		"kk..........",
		"kwk.........",
		"kwwk........",
		"kwwwk.......",
		"kwwwwk......",
		"kwwwwwk.....",
		"kwwwwwwk....",
		"kwwwwwwwk...",
		"kwwwwwwwwk..",
		"kwwwwwwwwwk.",
		"kwwwwwwkkkkk",
		"kwwwkwwk....",
		"kwwk.kwwk...",
		"kwk..kwwk...",
		"kk....kwwk..",
		"k.....kwwk..",
		".......kwwk.",
		"........kk..",
	}, map[rune]RGB{'k': cBlack, 'w': cWhite})
	return l
}

// Pointer is the hand cursor for links (16x19).
func Pointer() *Layer {
	l := NewLayer(16, 19)
	l.Sprite(0, 0, []string{
		".....kk.........",
		"....kwwk........",
		"....kwwk........",
		"....kwwk........",
		"....kwwkkk......",
		"....kwwkwwkkk...",
		"....kwwkwwkwwkk.",
		".kk.kwwkwwkwwkwk",
		"kwwkkwwwwwwwwkwk",
		"kwwwkwwwwwwwwwwk",
		".kwwkwwwwwwwwwwk",
		"..kwwwwwwwwwwwwk",
		"..kwwwwwwwwwwwk.",
		"...kwwwwwwwwwwk.",
		"...kwwwwwwwwwk..",
		"....kwwwwwwwwk..",
		"....kwwwwwwwwk..",
		"....kkkkkkkkkk..",
		"................",
	}, map[rune]RGB{'k': cBlack, 'w': cWhite})
	return l
}

// Favicon is a 32x32 CRT.
func Favicon() *Layer {
	l := NewLayer(32, 32)
	l.Rect(2, 4, 28, 22, cGrey)
	l.Rect(2, 4, 28, 1, cGreyLite)
	l.Rect(5, 7, 22, 15, Hex("#3a6ad8"))
	l.Rect(5, 17, 22, 5, Hex("#d8a868"))
	l.Rect(12, 11, 3, 6, cOrange)
	l.Rect(8, 26, 16, 3, cGreyDark)
	return l
}

// Speaker is the 16x16 sound toggle icon, on or off.
func Speaker(on bool) *Layer {
	l := NewLayer(16, 16)
	l.Sprite(0, 0, []string{
		"................",
		"......k.........",
		".....kk.........",
		"....kwk.........",
		"kkkkwwk.........",
		"kwwwwwk.........",
		"kwwwwwk.........",
		"kwwwwwk.........",
		"kwwwwwk.........",
		"kwwwwwk.........",
		"kkkkwwk.........",
		"....kwk.........",
		".....kk.........",
		"......k.........",
		"................",
		"................",
	}, map[rune]RGB{'k': cBlack, 'w': cWhite})
	if on {
		l.Sprite(9, 3, []string{"y.....", ".y..y.", ".y...y", ".y...y", ".y...y", ".y...y", ".y..y.", "y....."}, map[rune]RGB{'y': cYellow})
		l.Rect(8, 6, 1, 4, cYellow)
	} else {
		l.Line(9, 4, 15, 10, cRed)
		l.Line(15, 4, 9, 10, cRed)
		l.Line(10, 4, 15, 9, cRed)
		l.Line(14, 4, 9, 9, cRed)
	}
	return l
}

// Background is the 64x64 starfield tile behind every page.
func Background() *Layer {
	l := NewLayer(64, 64)
	l.Rect(0, 0, 64, 64, Hex("#140c2a"))
	seed := uint32(42)
	for i := 0; i < 26; i++ {
		seed = seed*1664525 + 1013904223
		x, y := int(seed>>8)%64, int(seed>>18)%64
		c := []RGB{Hex("#5a4a8a"), Hex("#8a7ac8"), Hex("#c8c0ff"), Hex("#3a2e6a")}[seed>>30]
		l.Set(x, y, c)
	}
	l.Sprite(40, 12, []string{".y.", "yWy", ".y."}, map[rune]RGB{'y': Hex("#c8a8ff"), 'W': cWhite})
	l.Sprite(10, 44, []string{".y.", "yWy", ".y."}, map[rune]RGB{'y': Hex("#8ac8ff"), 'W': cWhite})
	return l
}

// HotspotCSS positions the room's links over the image at Zoom.
func HotspotCSS() string {
	var b strings.Builder
	b.WriteString("/* Generated by `go run ./cmd/artgen` from internal/art/room.go. Do not edit. */\n")
	for _, h := range Hotspots {
		fmt.Fprintf(&b, ".hs-%s{left:%dpx;top:%dpx;width:%dpx;height:%dpx}\n", h.ID, h.X*Zoom, h.Y*Zoom, h.W*Zoom, h.H*Zoom)
	}
	return b.String()
}
