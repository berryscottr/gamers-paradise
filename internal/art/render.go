package art

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// Render produces every generated file, keyed by file name.
func Render() (map[string][]byte, error) {
	out := map[string][]byte{}
	add := func(name string, l *Layer) error {
		b, err := encode(l)
		if err != nil {
			return err
		}
		out[name] = b
		return nil
	}
	for _, p := range Phases {
		if err := add("room-"+p.Name+".png", Room(p).Scale(Zoom)); err != nil {
			return nil, err
		}
	}
	for _, t := range Titles {
		scale := 4
		if t.File == "title-home" {
			scale = 6
		}
		if err := add(t.File+".png", TitleArt(t, scale)); err != nil {
			return nil, err
		}
	}
	for name, l := range map[string]*Layer{
		"button.png":      Button(),
		"cursor.png":      Cursor(),
		"pointer.png":     Pointer(),
		"favicon.png":     Favicon(),
		"speaker-on.png":  Speaker(true).Scale(2),
		"speaker-off.png": Speaker(false).Scale(2),
		"bg.png":          Background(),
	} {
		if err := add(name, l); err != nil {
			return nil, err
		}
	}
	out["hotspots.css"] = []byte(HotspotCSS())
	out["font.js"] = []byte(FontJS())
	return out, nil
}

// encode writes a paletted PNG when the image has 256 colors or fewer
// (every room does, thanks to the banding in compose), else RGBA.
func encode(l *Layer) ([]byte, error) {
	rgba := l.Image()
	seen := map[color.RGBA]uint8{}
	var pal color.Palette
	transparent := false
	for i := 0; i < len(rgba.Pix); i += 4 {
		c := color.RGBA{rgba.Pix[i], rgba.Pix[i+1], rgba.Pix[i+2], rgba.Pix[i+3]}
		if c.A == 0 {
			transparent = true
			continue
		}
		if _, ok := seen[c]; !ok && len(pal) <= 256 {
			seen[c] = uint8(len(pal))
			pal = append(pal, c)
		}
	}
	if transparent {
		pal = append(color.Palette{color.RGBA{}}, pal...)
	}
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if len(pal) > 256 {
		err := enc.Encode(&buf, rgba)
		return buf.Bytes(), err
	}
	img := image.NewPaletted(rgba.Bounds(), pal)
	index := map[color.RGBA]uint8{}
	for i, c := range pal {
		index[c.(color.RGBA)] = uint8(i)
	}
	b := rgba.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := rgba.RGBAAt(x, y)
			if c.A == 0 {
				c = color.RGBA{}
			}
			img.SetColorIndex(x, y, index[c])
		}
	}
	err := enc.Encode(&buf, img)
	return buf.Bytes(), err
}
