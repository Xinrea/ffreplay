package ui

import (
	"fmt"
	"image"

	"github.com/Xinrea/ffreplay/internal/planner"
	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	layersPerPage = 5
	layerPitch    = 32
)

// Rows run front-to-back on screen, the opposite of document paint order.
// Keep row bounds, insertion markers and the eventual drop on one geometry.
type layerPage struct {
	top         image.Point
	first, rows int // document index of the top visible layer, visible row count
}

func (p layerPage) row(row int) image.Rectangle {
	y := p.top.Y + row*layerPitch
	return image.Rect(p.top.X, y, p.top.X+rightRail-32, y+layerPitch-2)
}

func (p layerPage) drop(at image.Point, source int) (boundary, target int, ok bool) {
	if p.rows == 0 || at.X < p.top.X || at.X >= p.row(0).Max.X || at.Y < p.top.Y-3 || at.Y > p.top.Y+p.rows*layerPitch+3 {
		return 0, 0, false
	}
	boundary = min(p.rows, max(0, (at.Y-p.top.Y+layerPitch/2)/layerPitch))
	// Insert at this gap, accounting for the source being removed first.
	target = p.first - boundary
	if source > target {
		target++
	}
	return boundary, target, true
}

type layerGesture struct {
	page   layerPage
	start  image.Point
	source int
	moving bool
}

func (b *StrategyBoard) updateLayerGesture() {
	d := b.layerDrag
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || !ebiten.IsFocused() {
		b.layerDrag = nil
		return
	}
	at := image.Pt(input.CursorPosition())
	diff := at.Sub(d.start)
	d.moving = d.moving || diff.X*diff.X+diff.Y*diff.Y >= 25
	if input.MouseButtonPressed(ebiten.MouseButtonLeft) {
		return
	}
	b.layerDrag = nil
	if !input.MouseButtonJustReleased(ebiten.MouseButtonLeft) {
		return
	}
	if !d.moving {
		if at.In(d.page.row(d.page.first - d.source)) {
			b.selectObject(d.source, ebiten.IsKeyPressed(ebiten.KeyShift))
		}
		return
	}
	if _, target, ok := d.page.drop(at, d.source); ok && target != d.source {
		b.moveLayer(d.source, target)
		b.status = fmt.Sprintf("%s 已移至第 %d 层，可撤销", b.editor.Step().Objects[target].Label, target+1)
	}
}

// Keep native button focus/keyboard activation, but align the grip, icon,
// name and depth in separate columns rather than padding a string with spaces.
type layerRow struct {
	*widget.Button
	glyph    *ebiten.Image
	depth    int
	selected bool
}

func (row *layerRow) Render(screen *ebiten.Image) {
	row.Button.Render(screen)
	r := row.GetWidget().Rect
	x, y := float32(r.Min.X), float32(r.Min.Y+r.Dy()/2)
	if row.selected {
		vector.FillRect(screen, x, y-10, 2, 20, boardGold, false)
	}
	for col := range 2 {
		for line := range 3 {
			vector.FillCircle(screen, x+9+float32(col*4), y+float32((line-1)*4), 1, boardMuted, true)
		}
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(20/float64(row.glyph.Bounds().Dx()), 20/float64(row.glyph.Bounds().Dy()))
	op.GeoM.Translate(float64(x)+22, float64(y)-10)
	screen.DrawImage(row.glyph, op)
	boardText(screen, fmt.Sprintf("%02d", row.depth), 12, float64(r.Max.X-16), float64(y), 28, 0, true, boardMuted)
}

// buildLayers returns the top of the inspector section below the visible rows.
func (b *StrategyBoard) buildLayers(x int) int {
	objects := b.editor.Step().Objects
	count := len(objects)
	pages := max(1, (count+layersPerPage-1)/layersPerPage)
	b.page = min(b.page, pages-1)
	p := layerPage{image.Pt(x, 195), count - 1 - b.page*layersPerPage, min(layersPerPage, count-b.page*layersPerPage)}
	b.label(x, 147, "对象图层", 16, boardGold)
	b.label(x+75, 151, fmt.Sprintf("%d 层", count), 12, boardMuted)
	if pages > 1 {
		pageButton := func(offset int, label string, delta int) {
			b.button(x+offset, 140, 28, label, false, func() { b.page += delta; b.rebuild = true },
				widget.ButtonOpts.Image(hudListRow(false))).GetWidget().Disabled = b.page+delta < 0 || b.page+delta >= pages
		}
		pageButton(122, "‹", -1)
		b.label(x+152, 150, fmt.Sprintf("%d/%d", b.page+1, pages), 12, boardMuted)
		pageButton(188, "›", 1)
	}
	if count == 0 {
		b.label(x, 176, "放置对象后，在这里调整层级", 13, boardMuted)
		return 218
	}
	b.label(x, 174, "从上到下显示 · 拖动调整顺序", 12, boardMuted)
	face := newEUIFace(14)
	for row := range p.rows {
		i := p.first - row
		obj := objects[i]
		r := p.row(row)
		btn := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(boardAt(r.Min.X, r.Min.Y, r.Dx(), r.Dy())...),
			widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.CursorHovered(input.CURSOR_POINTER)),
			widget.ButtonOpts.Image(hudListRow(b.selected[i])),
			widget.ButtonOpts.Text(shortLabel(obj.Label, 9), &face, &widget.ButtonTextColor{Idle: boardInk, Hover: boardGold}),
			widget.ButtonOpts.TextPadding(&widget.Insets{Left: 50, Right: 34}),
			widget.ButtonOpts.TextPosition(widget.TextPositionStart, widget.TextPositionCenter),
			widget.ButtonOpts.PressedHandler(func(*widget.ButtonPressedEventArgs) {
				if input.MouseButtonPressed(ebiten.MouseButtonLeft) {
					b.layerDrag = &layerGesture{page: p, source: i, start: image.Pt(input.CursorPosition())}
				}
			}),
			widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) {
				// Pointer clicks are resolved together with drags on release.
				if b.layerDrag == nil && !input.MouseButtonJustReleased(ebiten.MouseButtonLeft) {
					b.selectObject(i, ebiten.IsKeyPressed(ebiten.KeyShift))
				}
			}),
		)
		b.ui.Container.AddChild(&layerRow{btn, toolGlyph(planner.Tools[obj.Kind]), i + 1, b.selected[i]})
	}
	y := p.top.Y + p.rows*layerPitch + 6
	idx, single := b.single()
	b.button(x, y, 64, "上移", false, func() { b.layer(1) }).GetWidget().Disabled = !single || idx == count-1
	b.button(x+70, y, 64, "下移", false, func() { b.layer(-1) }).GetWidget().Disabled = !single || idx == 0
	b.button(x+152, y, 64, "删除", false, b.removeSelected).GetWidget().Disabled = len(b.selected) == 0
	return y + 48
}

func (b *StrategyBoard) drawLayerFeedback(screen *ebiten.Image) {
	d := b.layerDrag
	if d == nil || !d.moving {
		return
	}
	at := image.Pt(input.CursorPosition())
	source := d.page.row(d.page.first - d.source)
	vector.FillRect(screen, float32(source.Min.X), float32(source.Min.Y), float32(source.Dx()), float32(source.Dy()), hudDragShade, false)
	boundary, target, ok := d.page.drop(at, d.source)
	hint := "移出列表 · 松开取消"
	if ok {
		x, y := float32(d.page.top.X), float32(d.page.top.Y+boundary*layerPitch-1)
		vector.StrokeLine(screen, x+3, y, float32(source.Max.X), y, 2, boardGold, false)
		vector.FillCircle(screen, x+3, y, 3, boardGold, true)
		hint = fmt.Sprintf("松开移至第 %d 层", target+1)
		if target == d.source {
			hint = "保持当前层级"
		}
	}
	// Offset the badge to the left of the pointer to keep the gap visible.
	x := max(4, min(at.X-179, b.width-174))
	y := max(topBar, min(at.Y+16, b.height-bottomBar-48))
	drawHUDWindow(screen, image.Rect(x, y, x+170, y+46), hudTitle)
	boardText(screen, shortLabel(b.editor.Step().Objects[d.source].Label, 10), 14, float64(x+12), float64(y+14), 146, 0, false, boardInk)
	boardText(screen, hint, 12, float64(x+12), float64(y+32), 146, 0, false, boardGold)
}
