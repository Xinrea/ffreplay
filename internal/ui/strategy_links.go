package ui

import (
	"image/color"
	"math"

	"github.com/Xinrea/ffreplay/internal/planner"
	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var linkNames = []string{"普通线", "靠近", "远离", "负 / 负", "正 / 负", "正 / 正"}

type linkPoint struct {
	index int
	p     planner.Endpoint
}

func (b *StrategyBoard) linkPalette() {
	b.label(16, 333, "选择线型，再点两个端点", 14, boardInk)
	for i, style := range planner.LinkStyles {
		b.button(16+i%2*87, 365+i/2*36, 82, linkNames[i], b.linkTool == style, func() {
			b.resetSelection()
			b.linkTool = style
			b.panel = ""
			b.status = "请选择起点：对象或场地位置"
		})
	}
	b.label(16, 487, "点击已有连线可修改或删除", 13, boardMuted)
	n := len(b.editor.Step().Links)
	b.button(16, 507, 170, "选中下一条连线", false, func() {
		next := b.selectedLink%n + 1
		b.resetSelection()
		b.selectedLink = next
		if b.compact {
			b.panel = "对象"
		}
	}).GetWidget().Disabled = n == 0
}
func (b *StrategyBoard) linkInspector(x int) {
	l := &b.editor.Step().Links[b.selectedLink-1]
	b.label(x, 153, "连线属性", 17, boardGold)
	for i, style := range planner.LinkStyles {
		b.button(x+i%2*110, 190+i/2*36, 104, linkNames[i], l.Style == style, func() {
			if l.Style != style {
				b.editor.Checkpoint()
				l.Style = style
				b.rebuild = true
			}
		})
	}
	for i, name := range []string{"琥珀", "水蓝", "紫晶", "翠绿"} {
		b.button(x+i*55, 308, 50, name, l.Color == i, func() {
			if l.Color != i {
				b.editor.Checkpoint()
				l.Color = i
				b.rebuild = true
			}
		})
	}
	b.button(x, 354, 104, "删除连线", false, b.removeSelected)
}
func (b *StrategyBoard) linkPoint(x, y float64) linkPoint {
	s := b.editor.Step()
	i := pickLayer(s.Objects, x, y, len(s.Objects)-1)
	if i >= 0 {
		x, y = s.Objects[i].X, s.Objects[i].Y
	} else if b.snap {
		x, y = math.Round(x), math.Round(y)
	}
	return linkPoint{i, planner.Endpoint{X: x, Y: y}}
}
func (b *StrategyBoard) placeLink(x, y float64) {
	point := b.linkPoint(x, y)
	if b.linkStart == nil {
		b.linkStart = &point
		b.status = "请选择终点；Esc 取消"
		return
	}
	a := b.linkStart
	if i := b.editor.Connect(a.index, point.index, a.p, point.p, b.linkTool); i >= 0 {
		b.resetSelection()
		b.selectedLink = i + 1
		b.status = "已创建连线；点击线段可更改样式"
	} else {
		b.status = "请选择不同的端点，或检查连线数量"
	}
}
func (b *StrategyBoard) pickLink(x, y, unit float64) int {
	s := b.editor.Step()
	for i := len(s.Links) - 1; i >= 0; i-- {
		l := s.Links[i]
		if l.Game != nil && l.Game.Flags&1 == 0 {
			continue
		}
		a, c, _ := s.Position(l.Start)
		d, e, _ := s.Position(l.End)
		u, v := d-a, e-c
		t := planner.Clamp(((x-a)*u+(y-c)*v)/max(.000001, u*u+v*v), 0, 1)
		if math.Hypot(x-a-t*u, y-c-t*v) <= l.Stroke()/2+5/unit {
			return i
		}
	}
	return -1
}
func (b *StrategyBoard) drawBoardLink(dst *ebiten.Image, cx, cy, unit float64, i int) {
	s := b.editor.Step()
	l := s.Links[i]
	if l.Game != nil && l.Game.Flags&1 == 0 {
		return
	}
	x, y, _ := s.Position(l.Start)
	xx, yy, _ := s.Position(l.End)
	ink := boardInks[l.Color]
	if l.Game != nil && !l.Recolored() {
		ink = color.NRGBA{l.Game.RGBA[0], l.Game.RGBA[1], l.Game.RGBA[2], 255}
	}
	if l.Game != nil {
		ink.A = byte(255 * (100 - min(int(l.Game.RGBA[3]), 100)) / 100)
	}
	drawLink(dst, cx+x*unit, cy+y*unit, cx+xx*unit, cy+yy*unit, max(2, l.Stroke()*unit), l.Style, ink, b.selectedLink == i+1)
}
func (b *StrategyBoard) drawPendingLink(dst *ebiten.Image, cx, cy, unit float64) {
	if a := b.linkStart; a != nil {
		mx, my := input.CursorPosition()
		z := b.linkPoint((float64(mx)-cx)/unit, (float64(my)-cy)/unit)
		drawLink(dst, cx+a.p.X*unit, cy+a.p.Y*unit, cx+z.p.X*unit, cy+z.p.Y*unit, 2, b.linkTool, boardGold, false)
	}
}
func drawLink(dst *ebiten.Image, x, y, xx, yy, width float64, style string, ink color.NRGBA, selected bool) {
	stroke := func(a, b, c, d, w float64, col color.Color) {
		vector.StrokeLine(dst, float32(a), float32(b), float32(c), float32(d), float32(w), col, true)
	}
	if selected {
		stroke(x, y, xx, yy, width+4, boardGold)
	}
	stroke(x, y, xx, yy, width, ink)
	length := math.Hypot(xx-x, yy-y)
	if length < 1 || style == "line" {
		return
	}
	dx, dy := (xx-x)/length, (yy-y)/length
	for i, t := range []float64{.32, .68} {
		a, b := x+t*(xx-x), y+t*(yy-y)
		if style == "close" || style == "far" {
			sign := 1.
			if (i == 1) != (style == "far") {
				sign = -1
			}
			for _, side := range []float64{-1, 1} {
				stroke(a+dx*5*sign, b+dy*5*sign, a-dx*5*sign-dy*side*5, b-dy*5*sign+dx*side*5, 2.5, ink)
			}
		} else {
			vector.FillCircle(dst, float32(a), float32(b), 8, boardCanvas, true)
			vector.StrokeCircle(dst, float32(a), float32(b), 8, 1.5, ink, true)
			stroke(a-4, b, a+4, b, 2, ink)
			if style[i] == '+' {
				stroke(a, b-4, a, b+4, 2, ink)
			}
		}
	}
}
