package art

import "math"

// Room size in art pixels. The site shows it at 3x (960x600).
const (
	RoomW = 320
	RoomH = 200
	Zoom  = 3
)

// Phase is one lighting state of the room, picked by the visitor's clock.
type Phase struct {
	Name      string
	FromHour  int     // inclusive, local time
	Tint      RGB     // multiplies the room (255 = unchanged)
	SkyTop    RGB     // window gradient
	SkyBottom RGB     // window gradient
	Lights    float64 // 0: screens look switched off by daylight, 1: they glow
	CRTGlow   float64 // strength of the TV's light on the room
	Sun       bool    // sun (true) or moon and stars (false)
	SunY      int     // sun height inside the window
	Beam      bool    // morning light falling across the floor
}

// Phases in clock order. JavaScript picks one by hour (static/room.js);
// TestPhasesCoverDay keeps the two in sync.
var Phases = []Phase{
	{Name: "latenight", FromHour: 2, Tint: Hex("#2a2e58"), SkyTop: Hex("#03050f"), SkyBottom: Hex("#0b1030"), Lights: 1, CRTGlow: 0.55},
	{Name: "morning", FromHour: 5, Tint: Hex("#fff1dc"), SkyTop: Hex("#86c4ee"), SkyBottom: Hex("#f8d8a4"), Lights: 0.25, Sun: true, SunY: 30, Beam: true},
	{Name: "afternoon", FromHour: 11, Tint: Hex("#ffffff"), SkyTop: Hex("#4f98e2"), SkyBottom: Hex("#aed6f4"), Lights: 0.15, Sun: true, SunY: 8},
	{Name: "dusk", FromHour: 17, Tint: Hex("#e2b4b8"), SkyTop: Hex("#38286a"), SkyBottom: Hex("#f29a6a"), Lights: 0.6, CRTGlow: 0.15, Sun: true, SunY: 38},
	{Name: "night", FromHour: 21, Tint: Hex("#464c86"), SkyTop: Hex("#070b26"), SkyBottom: Hex("#1c2452"), Lights: 1, CRTGlow: 0.4},
}

// PhaseAt returns the phase for an hour of the day (0-23).
func PhaseAt(hour int) Phase {
	p := Phases[len(Phases)-1] // night wraps past midnight
	for _, ph := range Phases {
		if hour >= ph.FromHour {
			p = ph
		}
	}
	if hour < Phases[0].FromHour {
		return Phases[len(Phases)-1]
	}
	return p
}

// Hotspot is a clickable object in the room. Rect is in art pixels.
type Hotspot struct {
	ID    string
	Href  string // empty: not a link (a label only)
	Label string
	X, Y  int
	W, H  int
}

// Hotspots, back to front (later ones sit on top).
var Hotspots = []Hotspot{
	{ID: "window", Label: "The window. The sky follows your clock.", X: 8, Y: 14, W: 66, H: 56},
	{ID: "now", Href: "/now", Label: "The poster: right now. Hades, Persona and JoJo's.", X: 78, Y: 16, W: 50, H: 44},
	{ID: "street", Href: "/xbox#need-for-speed", Label: "Street racing poster", X: 134, Y: 18, W: 40, H: 32},
	{ID: "setlist", Href: "/xbox#guitar-hero", Label: "The setlist: One, on expert", X: 180, Y: 18, W: 40, H: 46},
	{ID: "shelf", Href: "/shelf", Label: "The shelf: games and minis", X: 222, Y: 36, W: 70, H: 32},
	{ID: "top8", Href: "/links", Label: "Top 8: sites I like, and my button", X: 234, Y: 68, W: 20, H: 16},
	{ID: "xbox", Href: "/xbox", Label: "The Xbox and the CRT: the Xbox years", X: 150, Y: 74, W: 56, H: 80},
	{ID: "guitar", Href: "/xbox#guitar-hero", Label: "Guitar controller", X: 204, Y: 92, W: 16, H: 60},
	{ID: "gameboy", Href: "/gameboy", Label: "Game Boy Color: grass starters, always", X: 6, Y: 92, W: 28, H: 28},
	{ID: "laptop", Href: "/", Label: "The laptop. It's showing this room.", X: 236, Y: 83, W: 28, H: 22},
	{ID: "box", Href: "/box", Label: "Mystery box: coming soon", X: 266, Y: 86, W: 24, H: 20},
	{ID: "door", Href: "/guestbook", Label: "The door: sign the guestbook", X: 296, Y: 44, W: 22, H: 92},
	{ID: "skateboard", Href: "/arcade/skate", Label: "Skateboard: play the skate game", X: 236, Y: 142, W: 38, H: 10},
	{ID: "beanbag", Href: "/arcade", Label: "Beanbag and controller: play a game", X: 146, Y: 154, W: 74, H: 40},
}

// palette
var (
	cWall      = Hex("#c9b48e")
	cWallLine  = Hex("#bba67f")
	cBase      = Hex("#6e5638")
	cFloor     = Hex("#6d4b3b")
	cFloorDot  = Hex("#5e4033")
	cFloorLite = Hex("#7a5646")
	cWood      = Hex("#7a5532")
	cWoodDark  = Hex("#5a3d22")
	cWoodLite  = Hex("#946a42")
	cWhite     = Hex("#f1ece0")
	cShade     = Hex("#d6cfbf")
	cBlack     = Hex("#1c1c1e")
	cDark      = Hex("#2e2e32")
	cGrey      = Hex("#8c8c88")
	cGreyDark  = Hex("#5c5c5a")
	cGreyLite  = Hex("#b4b4ae")
	cBlanket   = Hex("#2f5fa8")
	cBlanketL  = Hex("#4a7cc4")
	cBlanketD  = Hex("#234a86")
	cCurtain   = Hex("#3b5a3a")
	cCurtainL  = Hex("#4f744d")
	cGold      = Hex("#e8b84a")
	cRed       = Hex("#c23a32")
	cRedDark   = Hex("#7a1e1c")
	cGreen     = Hex("#4f9a2c")
	cGreenDark = Hex("#2c5a16")
	cGreenLite = Hex("#a4d46c")
	cPink      = Hex("#ee8cb4")
	cBlue      = Hex("#3a78d8")
	cYellow    = Hex("#f4d84a")
	cOrange    = Hex("#e8803a")
	cCardboard = Hex("#b8895a")
	cCardDark  = Hex("#8e6640")
	cBean      = Hex("#8a2a3a")
	cBeanLite  = Hex("#b04a5a")
	cBeanDark  = Hex("#621c2a")
	cRug       = Hex("#2f7a74")
	cRugEdge   = Hex("#c8a44a")
)

// scene holds the three layers of the room: base is lit by the phase tint;
// sky is always shown as is; lights glow in proportion to Phase.Lights.
type scene struct {
	base, sky, lights *Layer
}

// Room renders the bedroom for a phase. The laptop screen shows the room
// itself, two levels deep.
func Room(p Phase) *Layer {
	inner := NewLayer(RoomW, RoomH)
	inner.Rect(0, 0, RoomW, RoomH, cBlack)
	for i := 0; i < 2; i++ {
		inner = compose(drawScene(p, inner), p)
	}
	return inner
}

func compose(s scene, p Phase) *Layer {
	out := NewLayer(RoomW, RoomH)
	tint := func(c RGB) RGB {
		return RGB{
			clamp8(float64(c.R) * float64(p.Tint.R) / 255),
			clamp8(float64(c.G) * float64(p.Tint.G) / 255),
			clamp8(float64(c.B) * float64(p.Tint.B) / 255),
		}
	}
	glowAt := func(x, y int, cx, cy, r float64, col RGB, k float64, c RGB) RGB {
		d := math.Hypot(float64(x)-cx, (float64(y)-cy)*1.3)
		if d >= r {
			return c
		}
		f := (1 - d/r) * (1 - d/r) * k
		return RGB{clamp8(float64(c.R) + float64(col.R)*f), clamp8(float64(c.G) + float64(col.G)*f), clamp8(float64(c.B) + float64(col.B)*f)}
	}
	for y := 0; y < RoomH; y++ {
		for x := 0; x < RoomW; x++ {
			c, _ := s.base.At(x, y)
			c = tint(c)
			if p.CRTGlow > 0 {
				c = glowAt(x, y, 176, 98, 120, Hex("#6a9ae0"), p.CRTGlow, c)
				c = glowAt(x, y, 249, 94, 34, Hex("#9ab8e8"), p.CRTGlow*0.6, c)
			}
			if p.Beam && inBeam(x, y) {
				c = mix(c, Hex("#fff0c8"), 0.22)
			}
			if l, ok := s.lights.At(x, y); ok {
				c = mix(tint(l), l, p.Lights)
			}
			if sk, ok := s.sky.At(x, y); ok {
				c = sk
			}
			// band to a small palette so the PNGs stay tiny and look 2006
			out.Set(x, y, RGB{c.R &^ 3, c.G &^ 3, c.B &^ 3})
		}
	}
	return out
}

// inBeam is the patch of morning sun from the window across the floor.
func inBeam(x, y int) bool {
	if y < 138 || y > 196 {
		return false
	}
	t := float64(y-138) / 58
	left := 70 + t*40
	right := 118 + t*70
	return float64(x) >= left && float64(x) <= right
}

func drawScene(p Phase, laptop *Layer) scene {
	s := scene{NewLayer(RoomW, RoomH), NewLayer(RoomW, RoomH), NewLayer(RoomW, RoomH)}
	b := s.base

	// wall with striped wallpaper, then baseboard and carpet
	b.Rect(0, 0, RoomW, 134, cWall)
	for x := 4; x < RoomW; x += 8 {
		b.Rect(x, 0, 1, 134, cWallLine)
	}
	b.Rect(0, 0, 3, 134, cWallLine)
	b.Rect(0, 134, RoomW, 4, cBase)
	b.Rect(0, 134, RoomW, 1, cWoodLite)
	b.Rect(0, 138, RoomW, 62, cFloor)
	seed := uint32(7)
	for i := 0; i < 900; i++ {
		seed = seed*1664525 + 1013904223
		x, y := int(seed>>8)%RoomW, 138+int(seed>>20)%62
		if seed&1 == 0 {
			b.Set(x, y, cFloorDot)
		} else {
			b.Set(x, y, cFloorLite)
		}
	}

	// rug
	b.Rect(34, 160, 98, 34, cRugEdge)
	b.Rect(36, 162, 94, 30, cRug)
	for x := 40; x < 128; x += 6 {
		b.Rect(x, 176, 3, 1, cRugEdge)
	}

	window(s, p)
	nowPoster(b)
	streetPoster(b)
	setlist(b)
	shelf(s)
	desk(s, laptop)
	door(b)
	nightstand(b)
	bed(b)
	tv(s)
	guitar(b)
	beanbag(b)
	return s
}

func window(s scene, p Phase) {
	b := s.base
	// curtains behind the frame edges
	for _, x := range []int{6, 64} {
		b.Rect(x, 10, 12, 64, cCurtain)
		for i := 2; i < 12; i += 4 {
			b.Rect(x+i, 10, 1, 64, cCurtainL)
		}
	}
	b.Rect(4, 8, 76, 3, cWoodDark) // rod
	b.Rect(14, 16, 52, 50, cWhite) // frame
	b.Rect(12, 66, 56, 3, cShade)  // sill
	sky := s.sky
	for y := 19; y < 63; y++ {
		c := mix(p.SkyTop, p.SkyBottom, float64(y-19)/44)
		sky.Rect(17, y, 46, 1, c)
	}
	if p.Sun {
		sy := 19 + p.SunY
		sky.Ellipse(50, sy+4, 4, 4, Hex("#fff2b0"))
		if p.SunY < 20 { // clouds on a bright day
			sky.Rect(22, 30, 12, 3, cWhite)
			sky.Rect(25, 28, 6, 2, cWhite)
			sky.Rect(34, 44, 10, 2, cWhite)
		}
	} else {
		sky.Ellipse(48, 27, 3, 3, Hex("#f4f0d8"))
		sky.Ellipse(50, 26, 2, 2, p.SkyTop) // crescent bite
		for _, st := range [][2]int{{21, 23}, {28, 35}, {37, 22}, {56, 40}, {24, 50}, {44, 47}, {59, 24}, {33, 56}} {
			sky.Set(st[0], st[1], Hex("#f4f0d8"))
		}
		sky.Rect(17, 55, 46, 8, Hex("#0a0d1e")) // rooftops
		sky.Rect(22, 51, 8, 4, Hex("#0a0d1e"))
		sky.Rect(44, 49, 10, 6, Hex("#0a0d1e"))
		sky.Set(47, 51, Hex("#f6d870")) // a neighbor is up late too
	}
	// mullions are part of the room (lit by the phase), so cut them out of the sky
	sky.Clear(38, 19, 3, 44)
	sky.Clear(17, 39, 46, 3)
	b.Rect(38, 19, 3, 44, cWhite)
	b.Rect(17, 39, 46, 3, cWhite)
}

func nowPoster(b *Layer) {
	b.Rect(78, 16, 50, 44, cBlack)
	// Hades: a pomegranate over a three-headed dog
	b.Rect(80, 18, 15, 40, Hex("#3a0e12"))
	b.Sprite(83, 21, []string{
		"...gg...",
		"..rgrr..",
		".rrrrrr.",
		"rrRrrrrr",
		"rRrrrrrr",
		"rrrrrrrr",
		".rrrrrr.",
		"..rrrr..",
	}, map[rune]RGB{'r': Hex("#b82432"), 'R': Hex("#f06a6a"), 'g': Hex("#5a8a2a")})
	b.Sprite(81, 40, []string{
		".kk..kk..kk..",
		"kkkekkkekkke.",
		".kkk.kkk.kkk.",
		"..kkkkkkkkk..",
		"..kkkkkkkkkk.",
		"..kkkkkkkkkk.",
		"..k.k....k.k.",
		"..k.k....k.k.",
	}, map[rune]RGB{'k': Hex("#120406"), 'e': Hex("#ff5030")})
	b.Rect(80, 54, 15, 1, Hex("#e8803a"))
	// Persona: a tarot card and a mask
	b.Rect(96, 18, 15, 40, Hex("#14306e"))
	b.Rect(99, 21, 9, 14, cWhite)
	b.Frame(99, 21, 9, 14, Hex("#2a5ac8"))
	b.Sprite(101, 25, []string{"..y..", ".yyy.", "yyyyy", ".yyy.", "..y.."}, map[rune]RGB{'y': cGold})
	b.Sprite(98, 40, []string{
		".wwwwwww.",
		"wwwwwwwww",
		"wkkwwwkkw",
		"wwwwwwwww",
		".wwwrwww.",
		"..wwwww..",
		"...www...",
	}, map[rune]RGB{'w': cWhite, 'k': Hex("#0a1430"), 'r': Hex("#d84a4a")})
	// JoJo's: a dramatic pose and menacing letters
	b.Rect(112, 18, 14, 40, Hex("#5a2a7a"))
	b.Sprite(114, 22, []string{
		"...kk...",
		"...kk..k",
		"..kkkkk.",
		".k.kk...",
		"k..kk...",
		"...kk...",
		"..k..k..",
		"..k...k.",
		".k....k.",
		".k.....k",
	}, map[rune]RGB{'k': Hex("#f6d84a")})
	gogo := []string{"....##", "####..", "...#..", "...#..", "####.."}
	b.Sprite(113, 36, gogo, map[rune]RGB{'#': cPink})
	b.Sprite(119, 43, gogo, map[rune]RGB{'#': cPink})
	b.Sprite(113, 50, gogo, map[rune]RGB{'#': cPink})
	b.Rect(78, 16, 50, 1, cGold) // thumbtacks
	b.Set(79, 17, cRed)
	b.Set(126, 17, cRed)
}

func streetPoster(b *Layer) {
	b.Rect(134, 18, 40, 32, Hex("#0e0a24"))
	for y := 18; y < 34; y++ {
		b.Rect(134, y, 40, 1, mix(Hex("#0e0a24"), Hex("#5a1a6a"), float64(y-18)/16))
	}
	for _, bl := range [][4]int{{136, 26, 6, 14}, {143, 22, 5, 18}, {150, 28, 7, 12}, {159, 24, 5, 16}, {166, 29, 6, 11}} {
		b.Rect(bl[0], bl[1], bl[2], bl[3], Hex("#06040e"))
		for wy := bl[1] + 2; wy < bl[1]+bl[3]-1; wy += 3 {
			b.Set(bl[0]+1+(wy%2), wy, Hex("#f2c44a"))
		}
	}
	b.Rect(134, 40, 40, 10, Hex("#1a1a22")) // road
	b.Rect(134, 44, 40, 1, Hex("#3a3a44"))
	b.Rect(136, 39, 36, 1, Hex("#ff3aa8")) // neon
	b.Rect(134, 41, 18, 1, Hex("#3ae0ff"))
	b.Sprite(150, 41, []string{
		"...kkkkk....",
		"..kccccck...",
		"kkkkkkkkkkkh",
		"kkkkkkkkkkkk",
		".ww.....ww..",
	}, map[rune]RGB{'k': Hex("#2a62d8"), 'c': Hex("#9ad8ff"), 'h': Hex("#fff6a0"), 'w': cBlack})
	b.Rect(162, 43, 8, 1, Hex("#fff6a0")) // headlight streak
	b.Rect(140, 43, 9, 1, Hex("#3ae0ff")) // nitrous trail
	b.Frame(134, 18, 40, 32, cBlack)
}

func setlist(b *Layer) {
	b.Rect(180, 18, 40, 46, cWhite)
	b.Frame(180, 18, 40, 46, cShade)
	b.Rect(182, 20, 36, 6, cRed)
	b.Rect(184, 22, 32, 2, cGold)
	for i, y := 0, 30; y < 60; i, y = i+1, y+5 {
		if i == 3 {
			continue
		}
		b.Rect(184, y, 18+(i*7)%12, 1, cGreyLite)
		b.Rect(208, y, 8, 1, cGreyLite)
	}
	b.Rect(181, 43, 38, 9, Hex("#fff0a0"))
	b.Text(183, 44, "ONE", cRedDark)
	for i := 0; i < 5; i++ {
		b.Set(203+i*3, 47, cGold)
		b.Set(203+i*3, 46, cGold)
	}
	b.Set(200, 18, cRed) // tack
}

func shelf(s scene) {
	b := s.base
	b.Rect(222, 64, 70, 3, cWoodLite)
	b.Rect(222, 67, 70, 1, cWoodDark)
	b.Rect(226, 68, 2, 5, cWoodDark)
	b.Rect(286, 68, 2, 5, cWoodDark)
	cases := []RGB{Hex("#2a8a3a"), Hex("#1a1a1a"), Hex("#c8a020"), Hex("#7a1a1a"), Hex("#2a4a9a"), Hex("#e05a1a"), Hex("#4a4a4a")}
	for i, c := range cases {
		x := 224 + i*4
		h := 14 - (i%3)%2
		b.Rect(x, 64-h, 3, h, c)
		b.Rect(x, 64-h+2, 3, 1, cWhite)
	}
	// d20
	b.Sprite(253, 55, []string{
		"...rr...",
		"..rRRr..",
		".rRrrRr.",
		"rRRrrRRr",
		"rrRRRRrr",
		".rRrrRr.",
		"..rrrr..",
		"...rr...",
	}, map[rune]RGB{'r': Hex("#9a1a2a"), 'R': Hex("#e8505a")})
	// a tiny adventuring party: wizard, fighter, rogue
	b.Sprite(263, 54, []string{
		"..p.....g..",
		"..p..s..g..",
		".ppp.s.ggg.",
		".fff.sffff.",
		".ppp.ss.g..",
		".ppp.ss.g..",
		"..p..s..g..",
		".p.p.s.g.g.",
		".p.p.s.g.g.",
		"...........",
	}, map[rune]RGB{'p': Hex("#6a3ab0"), 'f': Hex("#e8c4a0"), 's': cGreyLite, 'g': Hex("#3a7a3a")})
	// laser sword on its stand
	b.Rect(277, 61, 8, 3, cDark)
	b.Rect(280, 55, 2, 6, cGreyLite)
	b.Set(280, 57, cRed)
	s.lights.Rect(280, 39, 2, 16, Hex("#7ac8ff"))
	s.lights.Rect(280, 39, 1, 16, Hex("#e6f6ff"))
}

func desk(s scene, laptop *Layer) {
	b := s.base
	b.Rect(226, 104, 66, 3, cWoodLite)
	b.Rect(226, 107, 66, 1, cWoodDark)
	b.Rect(228, 108, 3, 44, cWood)
	b.Rect(287, 108, 3, 44, cWood)
	b.Rect(262, 108, 28, 22, cWood)
	b.Frame(262, 108, 28, 22, cWoodDark)
	b.Rect(272, 112, 8, 1, cGold)
	b.Rect(272, 122, 8, 1, cGold)
	// Top 8 sticky note
	b.Rect(236, 70, 16, 13, Hex("#f6e27a"))
	b.Rect(236, 70, 16, 2, Hex("#e8c84a"))
	b.Text(238, 73, "8", cBlue)
	for y := 74; y < 82; y += 2 {
		b.Rect(244, y, 6, 1, Hex("#b09a40"))
	}
	// laptop: base, screen bezel, and the room on its screen
	b.Rect(234, 101, 30, 3, cGreyLite)
	b.Rect(234, 103, 30, 1, cGreyDark)
	b.Rect(237, 83, 24, 18, cGreyDark)
	sw, sh := 22, 14
	for y := 0; y < 16; y++ {
		for x := 0; x < sw; x++ {
			if y == 0 || y > sh {
				s.lights.Set(238+x, 84+y, cBlack)
				continue
			}
			s.lights.Set(238+x, 84+y, sample(laptop, x, y-1, sw, sh))
		}
	}
	// mystery box with hazard tape and a question mark
	b.Rect(268, 90, 20, 14, cCardboard)
	b.Rect(268, 88, 20, 2, cCardDark)
	b.Rect(266, 87, 5, 2, cCardboard)
	b.Rect(285, 87, 5, 2, cCardboard)
	for x := 268; x < 288; x++ {
		if (x/2)%2 == 0 {
			b.Set(x, 92, cBlack)
		} else {
			b.Set(x, 92, cYellow)
		}
	}
	b.Text(276, 95, "?", Hex("#4a2e14"))
	// skateboard under the desk
	b.Rect(238, 146, 34, 2, Hex("#3a2a6a"))
	b.Rect(237, 145, 2, 2, Hex("#3a2a6a"))
	b.Rect(271, 145, 2, 2, Hex("#3a2a6a"))
	b.Rect(250, 146, 10, 1, cOrange)
	for _, x := range []int{242, 244, 264, 266} {
		b.Set(x, 148, cWhite)
		b.Set(x, 149, cWhite)
	}
}

// sample box-filters src down to a w x h grid and returns cell (x, y).
func sample(src *Layer, x, y, w, h int) RGB {
	x0, x1 := x*src.W/w, (x+1)*src.W/w
	y0, y1 := y*src.H/h, (y+1)*src.H/h
	var r, g, bl, n int
	for j := y0; j < y1; j++ {
		for i := x0; i < x1; i++ {
			c, _ := src.At(i, j)
			r, g, bl, n = r+int(c.R), g+int(c.G), bl+int(c.B), n+1
		}
	}
	if n == 0 {
		return cBlack
	}
	return RGB{uint8(r / n), uint8(g / n), uint8(bl / n)}
}

func door(b *Layer) {
	b.Rect(294, 42, 26, 94, cWoodDark)
	b.Rect(296, 44, 22, 92, cWood)
	b.Frame(299, 64, 16, 28, cWoodDark)
	b.Frame(299, 98, 16, 32, cWoodDark)
	b.Rect(299, 50, 16, 7, cGold) // nameplate
	b.Rect(301, 53, 12, 1, Hex("#7a5a1a"))
	b.Ellipse(300, 96, 1, 1, cGold)
	// guestbook sheet with a pencil on a string
	b.Rect(301, 62, 12, 18, cWhite)
	for y := 66; y < 79; y += 2 {
		b.Rect(302, y, 10, 1, Hex("#9ab4d8"))
	}
	b.Rect(302, 66, 6, 1, cDark)
	b.Rect(302, 68, 8, 1, cDark)
	b.Set(307, 62, cRed)
	b.Line(312, 80, 314, 88, cGreyLite)
	b.Rect(313, 88, 2, 7, cYellow)
	b.Set(313, 95, Hex("#e8c4a0"))
}

func nightstand(b *Layer) {
	b.Rect(4, 118, 28, 3, cWoodLite)
	b.Rect(5, 121, 26, 30, cWood)
	b.Frame(7, 124, 22, 10, cWoodDark)
	b.Rect(16, 128, 4, 1, cGold)
	b.Frame(7, 137, 22, 10, cWoodDark)
	b.Rect(16, 141, 4, 1, cGold)
	// green Game Boy Color with a pink battery cover peeking out
	b.Sprite(9, 102, []string{
		".gggggggggg.",
		"gssssssssssg",
		"gsLLLLLLLLsg",
		"gsLLLLLLLLsg",
		"gsLLLLLLLLsg",
		"gsLLLLLLLLsg",
		"gssssssssssg",
		"gggggggggggg",
		"ggkgggggggag",
		"gkkkggggagGp",
		"ggkgggggggGp",
		"gggggggggggp",
		"gggggddggggp",
		"gggggggggggg",
		".gggggggggg.",
		"..........",
	}, map[rune]RGB{'g': cGreen, 'G': cGreenDark, 's': Hex("#2a2a34"), 'L': Hex("#b8cc8a"), 'k': cBlack, 'a': cRedDark, 'd': cGreenDark, 'p': cPink})
	// a cartridge with a green label
	b.Sprite(22, 110, []string{
		".ccccccc",
		"cccccccc",
		"cLLLLLLc",
		"cLwwwwLc",
		"cLLLLLLc",
		"cccccccc",
		"cdcdcdcc",
	}, map[rune]RGB{'c': cGreyLite, 'L': cGreen, 'w': cGreenLite, 'd': cGold})
	// Zzz: someone is asleep (status condition)
	b.Sprite(18, 90, []string{
		"......zzz",
		".......z.",
		"...zz.zzz",
		"....z....",
		"zz.zz....",
		".z.......",
		"zz.......",
	}, map[rune]RGB{'z': Hex("#6a7ac8")})
}

func bed(b *Layer) {
	b.Rect(32, 92, 6, 60, cWoodDark)
	b.Rect(33, 93, 4, 4, cWoodLite)
	b.Rect(38, 120, 102, 26, cWhite)
	b.Rect(38, 140, 102, 6, cShade)
	b.Ellipse(52, 117, 13, 6, cWhite) // pillow
	b.Rect(42, 121, 22, 2, cShade)
	b.Rect(66, 116, 74, 30, cBlanket)
	for x := 70; x < 140; x += 10 {
		b.Rect(x, 116, 2, 30, cBlanketL)
	}
	for y := 122; y < 146; y += 8 {
		b.Rect(66, y, 74, 1, cBlanketD)
	}
	b.Rect(66, 116, 74, 2, cBlanketL) // fold
	b.Rect(38, 146, 102, 6, cWood)
	b.Rect(38, 152, 3, 4, cWoodDark)
	b.Rect(137, 152, 3, 4, cWoodDark)
}

func tv(s scene) {
	b := s.base
	// stand with the Xbox on its shelf
	b.Rect(148, 120, 58, 3, cWoodLite)
	b.Rect(150, 123, 54, 31, cWood)
	b.Rect(152, 132, 50, 1, cWoodDark)
	b.Rect(156, 136, 34, 10, cBlack)
	b.Rect(156, 136, 34, 1, cDark)
	b.Rect(157, 143, 6, 1, cDark)
	b.Ellipse(173, 141, 2, 2, Hex("#2a6a2a"))
	b.Set(173, 141, Hex("#7ae05a"))
	b.Set(186, 138, Hex("#7ae05a")) // power light
	// CRT
	b.Rect(152, 76, 50, 44, cGrey)
	b.Rect(152, 76, 50, 2, cGreyLite)
	b.Rect(152, 114, 50, 6, cGreyDark)
	for x := 156; x < 180; x += 3 {
		b.Rect(x, 115, 1, 4, cDark)
	}
	b.Ellipse(194, 116, 1, 1, cDark)
	b.Ellipse(189, 116, 1, 1, cDark)
	b.Rect(156, 79, 42, 34, cDark)
	// on screen: an original skate game on a half-pipe
	l := s.lights
	for y := 81; y < 111; y++ {
		l.Rect(158, y, 38, 1, mix(Hex("#3a6ad8"), Hex("#9ad0f8"), float64(y-81)/30))
	}
	for x := 158; x < 196; x++ {
		d := float64(x-177) / 19
		top := 98 + int(10*(1-d*d))
		if top > 110 {
			top = 110
		}
		for y := 110; y > top; y-- {
			l.Set(x, y, Hex("#d8a868"))
		}
		l.Set(x, top, Hex("#f0d0a0"))
	}
	l.Rect(158, 110, 38, 1, Hex("#806040"))
	l.Sprite(164, 88, []string{
		"..k..",
		".ok..",
		"oook.",
		"..b.b",
		".b...",
		"wwwww",
	}, map[rune]RGB{'k': Hex("#ffd8a8"), 'o': cOrange, 'b': cBlue, 'w': cBlack})
	l.Rect(160, 83, 12, 1, cWhite) // score
	l.Rect(186, 83, 8, 1, cYellow)
	for _, c := range [][2]int{{158, 81}, {195, 81}, {158, 110}, {195, 110}} {
		l.Set(c[0], c[1], cDark) // rounded tube corners
	}
	for y := 82; y < 110; y += 2 {
		for x := 158; x < 196; x++ {
			if c, ok := l.At(x, y); ok {
				l.Set(x, y, mix(c, cBlack, 0.12)) // scanlines
			}
		}
	}
	// wired controller cable from the Xbox, around the beanbag
	b.Line(188, 144, 206, 152, cDark)
	b.Line(206, 152, 214, 168, cDark)
	b.Line(214, 168, 212, 180, cDark)
}

func guitar(b *Layer) {
	b.Rect(210, 93, 5, 5, cBlack) // headstock
	b.Rect(211, 98, 3, 30, Hex("#3a2a1a"))
	for i, c := range []RGB{cGreen, cRed, cYellow, cBlue, cOrange} {
		b.Set(212, 100+i*3, c)
	}
	// body: smaller upper bout, a waist, bigger lower bout
	b.Ellipse(212, 132, 6, 5, cRed)
	b.Ellipse(212, 143, 8, 7, cRed)
	b.Rect(206, 137, 13, 2, cRed)
	b.Ellipse(209, 141, 3, 4, Hex("#e8e4dc")) // pickguard
	b.Rect(207, 135, 10, 2, cBlack)           // strum bar
	b.Rect(214, 145, 3, 1, cGreyLite)
	b.Rect(214, 147, 3, 1, cGreyLite)
}

func beanbag(b *Layer) {
	b.Ellipse(174, 178, 26, 12, cBeanDark)
	b.Ellipse(173, 176, 25, 11, cBean)
	b.Ellipse(170, 165, 18, 8, cBean)
	b.Ellipse(165, 162, 8, 3, cBeanLite)
	b.Line(152, 172, 190, 170, cBeanDark)
	b.Line(171, 158, 174, 188, cBeanDark)
	// the controller, cable plugged in
	b.Sprite(204, 178, []string{
		".kkkkkkkkkkk.",
		"kkwkkkkkkkykk",
		"kwwwkkkkkgkbk",
		"kkwkkkkkkkrkk",
		"kkkkkkkkkkkkk",
		".kkk.....kkk.",
	}, map[rune]RGB{'k': Hex("#26262a"), 'w': cGreyLite, 'y': cYellow, 'g': cGreen, 'b': cBlue, 'r': cRed})
}
