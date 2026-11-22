package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/Xinrea/ffreplay/internal/planner"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	boardInks  = []color.NRGBA{{255, 184, 78, 255}, {90, 195, 243, 255}, {200, 150, 243, 255}, {121, 216, 164, 255}}
	boardIcons = map[string]*ebiten.Image{}
)

func boardIcon(path string) *ebiten.Image {
	if boardIcons[path] == nil {
		boardIcons[path] = scaleImage(texture.NewTextureFromFile(path), 30, 30)
	}
	return boardIcons[path]
}

// The palette uses the same silhouettes as the canvas. Keys here are tool IDs;
// texture-backed entries in boardIcons are keyed by asset path instead.
func toolGlyph(t planner.Tool) *ebiten.Image {
	if boardIcons[t.Key] == nil {
		glyph := ebiten.NewImage(22, 22)
		label := t.Label
		if t.Key == "text" {
			label = "T"
		}
		drawBoardObject(glyph, planner.Object{Kind: t.Key, Label: label, Radius: 9}, 11, 11, 1)
		boardIcons[t.Key] = glyph
	}
	return boardIcons[t.Key]
}

func boardText(dst *ebiten.Image, label string, size, x, y, maxWidth, angle float64, center bool, ink color.Color) {
	face := newEUIFace(size)
	w, h := text.Measure(label, face, 0)
	op := &text.DrawOptions{}
	if center {
		op.GeoM.Translate(-w/2, -h/2)
	} else {
		op.GeoM.Translate(0, -h/2)
	}
	if maxWidth > 0 && w > maxWidth {
		op.GeoM.Scale(maxWidth/w, maxWidth/w)
	}
	op.GeoM.Rotate(angle)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(ink)
	text.Draw(dst, label, face, op)
}

func (b *StrategyBoard) Draw(screen *ebiten.Image) {
	screen.Fill(boardCanvas)
	r := b.canvas()
	if r.Empty() {
		return
	}
	canvas := screen.SubImage(r).(*ebiten.Image)
	cx, cy, u := b.transform()
	b.drawArena(canvas, cx, cy, u)
	s := b.editor.Step()
	for i, j := range s.Layers() {
		if j >= 0 {
			b.drawBoardLink(canvas, cx, cy, u, j)
			continue
		}
		o := s.Objects[i]
		if o.Game != nil && o.Game.Flags&1 == 0 {
			continue
		}
		drawBoardObject(canvas, o, cx, cy, u)
		if len(o.Buffs) > 0 {
			gap := 3.
			// Leave the north rotation handle visible and clickable while editing.
			if b.selected[i] && len(b.selected) == 1 {
				gap += 20
			}
			drawBoardBuffs(canvas, o, cx+o.X*u, cy+o.Y*u-o.Radius*u-gap, o.Radius*u)
		}
	}
	b.drawPendingLink(canvas, cx, cy, u)
	for i := range b.selected {
		o := b.editor.Step().Objects[i]
		x, y, rad := float32(cx+o.X*u), float32(cy+o.Y*u), float32(o.Radius*u)
		vector.StrokeCircle(canvas, x, y, rad+2, 1.5, boardGold, true)
		if len(b.selected) == 1 {
			for _, kind := range boardHandles(o) {
				hx, hy := handlePosition(o, kind, u)
				sx, sy := float32(cx+hx*u), float32(cy+hy*u)
				if kind == "rotate" {
					vector.StrokeLine(canvas, x, y, sx, sy, 1, boardGold, true)
					vector.FillCircle(canvas, sx, sy, 4.5, boardGold, true)
				} else if kind == "shape" {
					vector.FillCircle(canvas, sx, sy, 5, boardInk, true)
				} else {
					vector.FillRect(canvas, sx-4, sy-4, 8, 8, boardGold, false)
				}
			}
		}
	}
	mx, my := input.CursorPosition()
	x, y := (float64(mx)-cx)/u, (float64(my)-cy)/u
	onBoard := image.Pt(mx, my).In(r) && (!b.compact || b.panel == "")
	if o, ok := b.placement(x, y); ok && onBoard {
		drawBoardObject(canvas, o, cx, cy, u)
	}
	b.drawWindows(screen)
	b.ui.Draw(screen)
	b.drawLayerFeedback(screen)
	fy := float64(b.height) - 21
	draft, ink := "未保存草稿", boardMuted
	if b.unsaved() {
		draft, ink = "有未保存更改", boardGold
	} else if b.hasDraft {
		draft = "草稿已保存"
	}
	vector.FillCircle(screen, 20, float32(fy), 3.5, ink, true)
	boardText(screen, draft, 13, 32, fy, 100, 0, false, ink)
	status := b.status
	if b.pending != "" {
		status = "放置" + planner.Tools[b.pending].Label + "：点击场地，Esc 取消"
	}
	w := b.width - 158
	if !b.compact && onBoard {
		w -= 170
		boardText(screen, fmt.Sprintf("X %5.1f    Y %5.1f m", x, y), 13, float64(b.width-158), fy, 150, 0, false, boardMuted)
	}
	boardText(screen, status, 14, 140, fy, float64(w), 0, false, boardMuted)
}

// Cache only the active floor. Source-atop masks the artwork to the selected
// arena, leaving the compass and the surrounding work surface legible.
func (b *StrategyBoard) arenaFloor(path string) *ebiten.Image {
	const size = 2 * planner.Extent * 25
	if b.floorArt == nil {
		b.floorArt = ebiten.NewImage(size, size)
	} else if b.floorKey == path && b.floorSquare == b.editor.Square {
		return b.floorArt
	}
	b.floorKey, b.floorSquare = path, b.editor.Square
	b.floorArt.Clear()
	if b.editor.Square {
		b.floorArt.Fill(boardFloor)
	} else {
		vector.FillCircle(b.floorArt, size/2, size/2, size/2, boardFloor, true)
	}
	img := texture.NewTextureFromFile(path)
	op := &ebiten.DrawImageOptions{GeoM: texture.CenterGeoM(img), Blend: ebiten.BlendSourceAtop}
	op.GeoM.Translate(size/2, size/2)
	op.ColorScale.ScaleAlpha(.54)
	b.floorArt.DrawImage(img, op)
	return b.floorArt
}

func (b *StrategyBoard) drawArena(dst *ebiten.Image, cx, cy, u float64) {
	x, y, r := float32(cx), float32(cy), float32(u*planner.Extent)
	if p := b.floors[b.editor.Step().Floor]; p != "" {
		img := b.arenaFloor(p)
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear, GeoM: texture.CenterGeoM(img)}
		// FFReplay floor art uses 25 pixels per in-game metre.
		op.GeoM.Scale(u/25, u/25)
		op.GeoM.Translate(cx, cy)
		dst.DrawImage(img, op)
	} else if b.editor.Square {
		vector.FillRect(dst, x-r, y-r, 2*r, 2*r, boardFloor, false)
	} else {
		vector.FillCircle(dst, x, y, r, boardFloor, true)
	}
	if b.editor.Grid {
		for i := -15; i <= 15; i += 5 {
			d := float32(float64(i) * u)
			l := r
			if !b.editor.Square {
				l = float32(math.Sqrt(float64(r*r - d*d)))
			}
			vector.StrokeLine(dst, x+d, y-l, x+d, y+l, 1, boardGrid, true)
			vector.StrokeLine(dst, x-l, y+d, x+l, y+d, 1, boardGrid, true)
		}
	}
	if b.editor.Square {
		vector.StrokeRect(dst, x-r, y-r, 2*r, 2*r, 2, boardRim, true)
	} else {
		vector.StrokeCircle(dst, x, y, r, 2, boardRim, true)
		for i := range 32 {
			s, c := math.Sincos(float64(i) * math.Pi / 16)
			length := 4.
			if i%4 == 0 {
				length = 9
			}
			vector.StrokeLine(dst, x+float32(s)*r, y-float32(c)*r, x+float32(s)*(r-float32(length)), y-float32(c)*(r-float32(length)), 1, boardRim, true)
		}
	}
	for i, label := range []string{"N", "E", "S", "W"} {
		s, c := math.Sincos(float64(i) * math.Pi / 2)
		boardText(dst, label, 16, cx+s*(float64(r)+19), cy-c*(float64(r)+19), 0, 0, true, boardGold)
	}
	vector.StrokeCircle(dst, x, y, 3, 1.5, boardRim, true)
}

func drawBoardObject(dst *ebiten.Image, o planner.Object, cx, cy, u float64) {
	x, y, r := cx+o.X*u, cy+o.Y*u, o.Radius*u
	angle := o.Angle * math.Pi / 180
	ink := boardInks[o.Color]
	if o.Game != nil {
		g := o.Game
		if o.UseGameInk() {
			ink = color.NRGBA{g.RGBA[0], g.RGBA[1], g.RGBA[2], 255}
		}
		ink.A = byte(255 * (100 - min(int(g.RGBA[3]), 100)) / 100)
	}
	if path := planner.Tools[o.Kind].Icon; path != "" {
		img := texture.NewTextureFromFile(path)
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		w, h := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())
		op.GeoM.Translate(-w/2, -h/2)
		if o.Game != nil {
			if o.Game.Flags&2 != 0 {
				op.GeoM.Scale(-1, 1)
			}
			if o.Game.Flags&4 != 0 {
				op.GeoM.Scale(1, -1)
			}
			op.ColorScale.ScaleAlpha(float32(ink.A) / 255)
		}
		op.GeoM.Scale(2*r/w, 2*r/h)
		op.GeoM.Rotate(angle)
		op.GeoM.Translate(x, y)
		dst.DrawImage(img, op)
		if o.Color != 0 {
			vector.StrokeCircle(dst, float32(x), float32(y), float32(r), 2, ink, true)
		}
		return
	}
	if o.Kind == "text" {
		boardText(dst, o.Label, r, x, y, 2*r, angle, true, ink)
		return
	}
	xf, yf, rf := float32(x), float32(y), float32(r)
	var p vector.Path
	switch o.Kind {
	case "circle", "donut":
		if sweep := o.Aperture(); sweep < 360 {
			mid, half := angle-math.Pi/2, sweep*math.Pi/360
			lo, hi := float32(mid-half), float32(mid+half)
			p.Arc(xf, yf, rf, lo, hi, vector.Clockwise)
			p.Arc(xf, yf, rf*float32(o.ShapeValue()), hi, lo, vector.CounterClockwise)
			p.Close()
			break
		}
		p.Arc(xf, yf, rf, 0, 2*math.Pi, vector.Clockwise)
		p.Close()
		if o.Kind == "donut" {
			inner := rf * float32(o.ShapeValue())
			p.MoveTo(xf+inner, yf)
			p.Arc(xf, yf, inner, 0, 2*math.Pi, vector.Clockwise)
			p.Close()
		}
	case "cone":
		p.MoveTo(xf, yf)
		half := o.ShapeValue() * math.Pi / 360
		p.Arc(xf, yf, rf, float32(angle-math.Pi/2-half), float32(angle-math.Pi/2+half), vector.Clockwise)
		p.Close()
	case "rect", "arrow":
		w := o.ShapeValue()
		pts := [][2]float64{{-w, -1}, {w, -1}, {w, 1}, {-w, 1}}
		if o.Kind == "arrow" {
			pts = [][2]float64{{0, -1}, {.32, -.45}, {.10, -.5}, {.10, 1}, {-.10, 1}, {-.10, -.5}, {-.32, -.45}}
		}
		s, c := math.Sincos(angle)
		for i, pt := range pts {
			px, py := float32(x+r*(pt[0]*c-pt[1]*s)), float32(y+r*(pt[0]*s+pt[1]*c))
			if i == 0 {
				p.MoveTo(px, py)
			} else {
				p.LineTo(px, py)
			}
		}
		p.Close()
	}
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(ink)
	fill := *op
	fill.ColorScale.ScaleAlpha(.38)
	vector.FillPath(dst, &p, &vector.FillOptions{FillRule: vector.FillRuleEvenOdd}, &fill)
	vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: 2}, op)
}
