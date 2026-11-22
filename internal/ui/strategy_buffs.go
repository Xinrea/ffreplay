package ui

import (
	"fmt"
	"slices"

	"github.com/Xinrea/ffreplay/internal/planner"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// Keep search and pagination while editing the list. Render only the current
// page; the full catalogue is metadata, not hundreds of scaled GPU images.
type boardBuffPanel struct {
	query string
	page  int
}

const buffPageSize = 36

var buffThumbnails = map[int]*ebiten.Image{}

func buffThumb(icon planner.BuffIcon) *ebiten.Image {
	if thumb := buffThumbnails[icon.ID]; thumb != nil {
		return thumb
	}
	thumb := scaleImage(texture.NewAbilityTexture(icon.Asset), 21, 28)
	buffThumbnails[icon.ID] = thumb
	return thumb
}

func buffCell(icon planner.BuffIcon, active bool, x, y, w int, click func()) *widget.Button {
	bg := boardSlot
	if active {
		bg = boardSlotActive
	}
	face := newEUIFace(14)
	opts := boardAt(x, y, w, 34)
	opts = append(opts, widget.WidgetOpts.ToolTip(widget.NewTextToolTip(fmt.Sprintf("%s · %d", icon.Name, icon.ID), &face, boardInk, hudTooltip)))
	return widget.NewButton(widget.ButtonOpts.Image(bg), widget.ButtonOpts.Graphic(&widget.GraphicImage{Idle: buffThumb(icon)}),
		widget.ButtonOpts.WidgetOpts(opts...), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { click() }))
}

func (b *StrategyBoard) buildBuffPanel(x, index int) {
	panel := b.buffEditor
	o := b.editor.Step().Objects[index]
	toggle := func(id int) {
		if b.editor.ToggleBuff(index, id) {
			b.rebuild = true
		}
	}
	b.label(x, 80, "头顶 Buff", 18, boardGold)
	b.button(x+152, 74, 64, "返回属性", false, func() { b.buffEditor = nil; b.rebuild = true })
	b.label(x, 120, fmt.Sprintf("%s   %d / %d", shortLabel(o.Label, 10), len(o.Buffs), planner.MaxBuffIcons), 15, boardGold)
	for i, id := range o.Buffs {
		icon, _ := planner.LookupBuffIcon(id)
		b.ui.Container.AddChild(buffCell(icon, true, x+(i%6)*36, 151+(i/6)*36, 33, func() { toggle(id) }))
	}
	if len(o.Buffs) == 0 {
		b.label(x, 158, "尚未添加", 14, boardMuted)
	}
	b.label(x, 231, "按添加顺序显示，点击图标移除", 13, boardMuted)
	b.label(x, 269, "状态图标库", 17, boardGold)

	grid := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(boardAt(x, 349, 216, 216)...))
	b.ui.Container.AddChild(grid)
	face := newEUIFace(14)
	pageText := widget.NewText(widget.TextOpts.Text("", &face, boardMuted), widget.TextOpts.WidgetOpts(boardAt(x+46, 586, 0, 0)...))
	b.ui.Container.AddChild(pageText)
	var refresh func()
	prev := b.button(x, 579, 34, "‹", false, func() { panel.page--; refresh() })
	next := b.button(x+182, 579, 34, "›", false, func() { panel.page++; refresh() })
	refresh = func() {
		grid.RemoveChildren()
		matches := planner.FindBuffIcons(panel.query)
		pages := max(1, (len(matches)+buffPageSize-1)/buffPageSize)
		panel.page = max(0, min(panel.page, pages-1))
		prev.GetWidget().Disabled = panel.page == 0
		next.GetWidget().Disabled = panel.page+1 == pages
		pageText.Label = fmt.Sprintf("%d / %d 页 · %d 项", panel.page+1, pages, len(matches))
		if len(matches) == 0 {
			grid.AddChild(widget.NewText(widget.TextOpts.Text("没有匹配的图标，试试其他名称或 ID", &face, boardMuted)))
			return
		}
		start := panel.page * buffPageSize
		for i, icon := range matches[start:min(start+buffPageSize, len(matches))] {
			active := slices.Contains(o.Buffs, icon.ID)
			cell := buffCell(icon, active, (i%6)*36, (i/6)*36, 33, func() { toggle(icon.ID) })
			cell.GetWidget().Disabled = !active && len(o.Buffs) >= planner.MaxBuffIcons
			grid.AddChild(cell)
		}
	}
	search := widget.NewTextInput(widget.TextInputOpts.Face(&face), widget.TextInputOpts.Image(boardInput),
		widget.TextInputOpts.Color(boardInputInk), widget.TextInputOpts.Padding(widget.NewInsetsSimple(6)),
		widget.TextInputOpts.Placeholder("搜索中文名称或图标 ID"),
		widget.TextInputOpts.WidgetOpts(boardAt(x, 300, 216, 32)...),
		widget.TextInputOpts.ChangedHandler(func(a *widget.TextInputChangedEventArgs) {
			if a.InputText != panel.query {
				panel.query, panel.page = a.InputText, 0
				refresh()
			}
		}))
	search.SetText(panel.query)
	b.ui.Container.AddChild(search)
	refresh()
}

// Status artwork stays upright even if the player token rotates.
func drawBoardBuffs(dst *ebiten.Image, o planner.Object, x, bottom, r float64) {
	h := planner.Clamp(r*1.4, 20, 32)
	w := h * BuffWidth / BuffHeight
	const gap = 2.
	width := float64(len(o.Buffs))*(w+gap) - gap
	bounds := dst.Bounds()
	left := planner.Clamp(x-width/2, float64(bounds.Min.X), float64(bounds.Max.X)-width)
	top := max(bottom-h, float64(bounds.Min.Y))
	for i, id := range o.Buffs {
		icon, ok := planner.LookupBuffIcon(id)
		if !ok {
			continue
		}
		img := texture.NewAbilityTexture(icon.Asset)
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Scale(w/float64(img.Bounds().Dx()), h/float64(img.Bounds().Dy()))
		op.GeoM.Translate(left+float64(i)*(w+gap), top)
		dst.DrawImage(img, op)
	}
}
