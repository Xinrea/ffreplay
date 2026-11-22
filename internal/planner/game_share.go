package planner

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/Xinrea/ffreplay/internal/gameboard"
)

// Game design coordinates are tenths of a pixel on a 512 x 384 board.
// Use one scale on both axes: 40 metres spans the game's full width.
const units = 128.
const circleRadius = 20. * 29 / 30

var gameKinds = func() map[uint16]string {
	m := map[uint16]string{9: "circle", 10: "cone", 11: "rect", 17: "donut", 94: "arrow", 100: "text", 101: "Viper", 102: "Pictomancer"}
	for i, k := range []string{"Paladin", "Monk", "Warrior", "Dragoon", "Bard", "WhiteMage", "BlackMage", "Summoner", "Scholar", "Ninja", "Machinist", "DarkKnight", "Astrologian", "Samurai", "RedMage", "", "Gunbreaker", "Dancer", "Reaper", "Sage"} {
		if k != "" {
			m[uint16(27+i)] = k
		}
	}
	for i, k := range []string{"A", "B", "C", "D", "1", "2", "3", "4"} {
		m[uint16(79+i)] = k
	}
	for i, k := range map[uint16]string{48: "MT", 49: "ST", 51: "H1", 52: "H2", 54: "D1", 55: "D2", 56: "D3", 57: "D4"} {
		m[i] = k
	}
	for _, g := range gameIcons {
		m[g.id] = g.tool.Key
	}
	return m
}()

func gameKind(id uint16) string { return gameKinds[id] }
func (o Object) UseGameInk() bool {
	return o.Game != nil && (o.Kind == "rect" || o.Kind == "text") && nearestInk(o.Game.RGBA) == o.Color
}

var gameIDs = func() map[string]uint16 {
	m := map[string]uint16{"boss": 64}
	for id, k := range gameKinds {
		m[k] = id
	}
	return m
}()

var gameInks = [4][4]byte{{255, 184, 78, 0}, {90, 195, 243, 0}, {200, 150, 243, 0}, {121, 216, 164, 0}}

func nearestInk(c [4]byte) int {
	best, d := 0, math.MaxFloat64
	for i, v := range gameInks {
		r, g, b := float64(v[0])-float64(c[0]), float64(v[1])-float64(c[1]), float64(v[2])-float64(c[2])
		if n := r*r + g*g + b*b; n < d {
			d, best = n, i
		}
	}
	return best
}
func gameRadius(id uint16) float64 {
	switch {
	case id == 9 || id == 10 || id == 17:
		return circleRadius
	case id == 100:
		return 30 * 5 / units
	case id >= 79 && id <= 86 || id == 94:
		return 1.875
	case strings.HasPrefix(gameKind(id), "game-"):
		return Tools[gameKind(id)].Radius
	default:
		return 1.25
	}
}
func gamePoint(x, y int16) Endpoint {
	return Endpoint{X: (float64(x) - 2560) / units, Y: (float64(y) - 1920) / units}
}
func gameCoord(x, y float64) (int16, int16) {
	return int16(math.Round(x*units + 2560)), int16(math.Round(y*units + 1920))
}
func normAngle(a float64) float64 { return math.Remainder(a, 360) }

// Game fans use the centre of their unrotated bounding box, then rotate
// clockwise from north. FFReplay uses the apex and the bisector instead.
func fanOffset(radius, aperture, angle float64) (float64, float64) {
	theta := aperture * math.Pi / 180
	l, r, t, b := math.Min(0, math.Sin(theta)), math.Max(0, math.Sin(theta)), -1., math.Max(0, -math.Cos(theta))
	if aperture >= 90 {
		r = 1
	}
	if aperture >= 180 {
		b = 1
	}
	if aperture >= 270 {
		l = -1
	}
	x, y := (l+r)*radius/2, (t+b)*radius/2
	s, c := math.Sincos(angle * math.Pi / 180)
	return x*c - y*s, x*s + y*c
}

func fromGame(g gameboard.Object) (Object, bool, error) {
	k := gameKind(g.Type)
	t, ok := Tools[k]
	if !ok {
		return Object{}, false, fmt.Errorf("暂不支持游戏对象 ID %d；未导入", g.Type)
	}
	p := gamePoint(g.X, g.Y)
	o := Object{Kind: k, Label: t.Label, X: p.X, Y: p.Y, Angle: normAngle(float64(g.Angle)), Radius: float64(g.Scale) * gameRadius(g.Type) / 100, Game: &g}
	approx := false
	switch k {
	case "rect":
		w, h := float64(g.Params[0])*5/units, float64(g.Params[1])*5/units
		if w > h {
			w, h = h, w
			o.Angle = normAngle(o.Angle + 90)
		}
		o.Radius = h
		o.Shape = Clamp(w/max(.001, h), .05, 1)
		o.Color = nearestInk(g.RGBA)
	case "text":
		if g.Text != "" {
			o.Label = g.Text
		}
		if len([]rune(o.Label)) > 60 {
			return Object{}, false, errors.New("游戏文字超过 60 字")
		}
		o.Color = nearestInk(g.RGBA)
	case "cone", "donut":
		aperture := float64(g.Params[0])
		if aperture <= 0 || aperture > 360 {
			aperture = 360
		}
		dx, dy := fanOffset(o.Radius*30/29, aperture, o.Angle)
		o.X -= dx
		o.Y -= dy
		o.Angle = normAngle(o.Angle + aperture/2)
		if k == "cone" {
			o.Shape = aperture
		} else {
			o.Shape = Clamp(float64(g.Params[1])/256*30/29, .1, .9)
			approx = o.Shape != float64(g.Params[1])/256*30/29
		}
	}
	approx = approx || math.Abs(o.X) > Extent || math.Abs(o.Y) > Extent || o.Radius < .5 || o.Radius > Extent
	o.X, o.Y, o.Radius = Clamp(o.X, -Extent, Extent), Clamp(o.Y, -Extent, Extent), Clamp(o.Radius, .5, Extent)
	return o, approx, nil
}

func ImportGame(code string) (Plan, string, error) {
	b, err := gameboard.Decode(code)
	if err != nil {
		return Plan{}, "", err
	}
	p := New()
	p.Square, p.Grid = gameArena(b.Background)
	s := &p.Steps[0]
	s.GameBackground = b.Background
	if b.Name != "" {
		if len([]rune(b.Name)) > 60 {
			return Plan{}, "", errors.New("板名过长")
		}
		s.Name = b.Name
	}
	note := ""
	for _, g := range slices.Backward(b.Objects) {
		if g.Type == 12 {
			a, z := gamePoint(g.X, g.Y), gamePoint(int16(g.Params[0]), int16(g.Params[1]))
			if math.Abs(a.X) > Extent || math.Abs(a.Y) > Extent || math.Abs(z.X) > Extent || math.Abs(z.Y) > Extent {
				return Plan{}, "", errors.New("游戏线段超出可编辑场地")
			}
			s.Links = append(s.Links, Link{Start: a, End: z, Style: "line", Color: nearestInk(g.RGBA), Width: Clamp(float64(g.Params[2])*10/units, .05, Extent), Game: &g, Above: len(s.Objects)})
			continue
		}
		o, approx, err := fromGame(g)
		if err != nil {
			return Plan{}, "", err
		}
		if approx {
			note = "；部分尺寸、扇环或边缘位置近似显示"
		}
		s.Objects = append(s.Objects, o)
	}
	return p, note, nil
}

func gameArena(bg uint16) (square, grid bool) { return bg != 3 && bg != 6, bg >= 2 && bg <= 4 }

func toGame(o Object) (gameboard.Object, error) {
	id := gameIDs[o.Kind]
	if id == 0 {
		return gameboard.Object{}, fmt.Errorf("游戏不支持 %s", o.Label)
	}
	g := gameboard.Object{Type: id, Flags: 1, RGBA: [4]byte{255, 255, 255, 0}, Scale: 100}
	var orig Object
	if o.Game != nil {
		g = *o.Game
		var err error
		orig, _, err = fromGame(g)
		if err != nil {
			return g, err
		}
	}
	geometry := o.Game == nil || orig.Radius != o.Radius || orig.Shape != o.Shape || orig.Angle != o.Angle
	x, y, angle := o.X, o.Y, o.Angle
	if o.Kind == "cone" || o.Kind == "donut" {
		aperture := o.Aperture()
		if o.Kind == "cone" {
			aperture = max(10, math.Round(o.ShapeValue()/10)*10)
		}
		angle = normAngle(angle - aperture/2)
		dx, dy := fanOffset(o.Radius*30/29, aperture, angle)
		x += dx
		y += dy
		if geometry {
			g.Params[0] = uint16(aperture)
			if o.Kind == "donut" {
				g.Params[1] = uint16(math.Round(o.ShapeValue() * 256 * 29 / 30))
			}
		}
	}
	if o.Game == nil || orig.X != o.X || orig.Y != o.Y || geometry {
		if o.Game != nil && !geometry {
			g.X += int16(math.Round((o.X - orig.X) * units))
			g.Y += int16(math.Round((o.Y - orig.Y) * units))
		} else {
			g.X, g.Y = gameCoord(x, y)
		}
	}
	if geometry {
		g.Angle = int16(math.Round(normAngle(angle)))
		g.Scale = byte(Clamp(math.Round(100*o.Radius/gameRadius(id)), 10, 200))
		switch o.Kind {
		case "rect":
			g.Scale = 100
			g.Params[0] = uint16(math.Round(o.Radius * o.ShapeValue() * units / 5))
			g.Params[1] = uint16(math.Round(o.Radius * units / 5))
		case "text":
			g.Scale = 100
		}
		if id == 15 && g.Params[1] == 0 {
			g.Params[1] = 1
		}
		if id == 110 && o.Game == nil {
			g.Params[0], g.Params[1] = 1, 1
		}
	}
	if o.Kind == "text" && (o.Game == nil || orig.Label != o.Label) {
		g.Text = o.Label
	}
	if (id == 11 || id == 100) && (o.Game == nil || orig.Color != o.Color) {
		copy(g.RGBA[:3], gameInks[o.Color][:3])
	}
	return g, nil
}

func ExportGame(p Plan, step int) (string, error) {
	if step < 0 || step >= len(p.Steps) {
		return "", errors.New("无此步骤")
	}
	s := p.Steps[step]
	bg := s.GameBackground
	square, grid := gameArena(bg)
	if bg == 0 || square != p.Square || grid != p.Grid {
		bg = 6
		if p.Square {
			bg = 7
		}
		if p.Grid {
			bg -= 3
		}
	}
	b := gameboard.Board{Name: s.Name, Background: bg}
	for oi, li := range s.Layers() {
		if oi >= 0 {
			o := s.Objects[oi]
			g, err := toGame(o)
			if err != nil {
				return "", err
			}
			b.Objects = append(b.Objects, g)
			continue
		}
		l := s.Links[li]
		x, y, ok := s.Position(l.Start)
		xx, yy, ok2 := s.Position(l.End)
		if !ok || !ok2 {
			return "", errors.New("连线缺少端点")
		}
		g := gameboard.Object{Type: 12, Flags: 1, Scale: 100, RGBA: gameInks[l.Color]}
		if l.Game != nil {
			g = *l.Game
			if nearestInk(g.RGBA) != l.Color {
				copy(g.RGBA[:3], gameInks[l.Color][:3])
			}
		}
		startX, startY := gameCoord(x, y)
		ex, ey := gameCoord(xx, yy)
		if l.Game == nil || g.X != startX || g.Y != startY || g.Params[0] != uint16(ex) || g.Params[1] != uint16(ey) {
			g.X, g.Y = startX, startY
			g.Params[0], g.Params[1] = uint16(ex), uint16(ey)
			g.Angle = int16(math.Round(normAngle(-math.Atan2(yy-y, xx-x)*180/math.Pi - 90)))
		}
		if l.Game == nil || l.Width != Clamp(float64(g.Params[2])*10/units, .05, Extent) {
			g.Params[2] = uint16(max(1., math.Round(l.Stroke()*units/10)))
		}
		b.Objects = append(b.Objects, g)
	}
	slices.Reverse(b.Objects)
	return gameboard.Encode(b)
}
