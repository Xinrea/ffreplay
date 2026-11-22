package ui

import (
	"image"
	"image/color"

	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

// The strategy board uses the game's dark window theme. Keep diagram inks
// separate: an AoE colour is part of a saved strategy, not window chrome.
var (
	boardInk      = color.NRGBA{243, 239, 228, 255}
	boardMuted    = color.NRGBA{197, 191, 177, 255}
	boardGold     = color.NRGBA{252, 222, 157, 255}
	boardCanvas   = color.NRGBA{23, 24, 26, 255}
	boardDisabled = color.NRGBA{150, 147, 139, 255}
	hudDragShade  = color.NRGBA{28, 26, 22, 150}
	boardFloor    = color.NRGBA{54, 55, 56, 255}
	boardGrid     = color.NRGBA{85, 84, 80, 255}
	boardRim      = color.NRGBA{182, 165, 129, 255}

	hudWindow  = hudBevel(9, 0x8e887c, 0x514e48, 0x43413e, 0x292929)
	hudTitle   = hudBevel(9, 0xb9af97, 0x625d53, 0x55524b, 0x343331)
	hudInset   = toEUINineSlice(messageTextureAtlas.GetNineSlice("input_bg.png"))
	hudTooltip = toEUINineSlice(messageTextureAtlas.GetNineSlice("message_bg.png"))

	boardButton, boardActive   = hudButtons()
	boardSlot, boardSlotActive = hudSlots()
	boardInput                 = &widget.TextInputImage{Idle: hudInset, Disabled: hudInset,
		Highlight: euiimage.NewNineSliceColor(color.NRGBA{188, 155, 85, 128})}
	boardInputInk = &widget.TextInputColor{Idle: boardInk, Disabled: boardDisabled, Caret: boardGold, DisabledCaret: boardMuted}
	menuSelected  = euiimage.NewNineSliceColor(hudColor(0x655941))
	menuHover     = euiimage.NewNineSliceColor(hudColor(0x53514b))
	menuRow       = &widget.ButtonImage{Idle: transparentNineSlice(), Hover: menuHover, Pressed: menuSelected}
	menuCurrent   = &widget.ButtonImage{Idle: menuSelected, Hover: euiimage.NewNineSliceColor(hudColor(0x77694c)), Pressed: menuSelected}
)

func hudListRow(selected bool) *widget.ButtonImage {
	if selected {
		return menuCurrent
	}
	return menuRow
}

func hudListParams(st propInspectorStyle) (*widget.ButtonParams, *widget.ListParams) {
	btn, list := propListButtonParams(st), propListParams(st)
	btn.Image = boardButton
	btn.TextColor = &widget.ButtonTextColor{Idle: boardInk, Disabled: boardDisabled}
	btn.MinSize = &image.Point{X: 170, Y: 28}
	list.MinSize = btn.MinSize
	list.ScrollContainerImage.Idle = hudWindow
	list.EntryColor = &widget.ListEntryColor{
		Unselected: boardInk, Selected: boardGold, FocusedBackground: hudColor(0x53514b),
		SelectedBackground: hudColor(0x655941), SelectedFocusedBackground: hudColor(0x77694c),
		DisabledUnselected: boardDisabled, DisabledSelected: boardDisabled,
	}
	list.Slider.TrackImage = &widget.SliderTrackImage{Idle: euiimage.NewNineSliceColor(hudColor(0x252523))}
	list.Slider.HandleImage = &widget.ButtonImage{Idle: euiimage.NewNineSliceColor(hudColor(0x898578)), Hover: menuHover, Pressed: menuSelected}
	return btn, list
}

func (b *StrategyBoard) hudToggle(x, y int, label string, value bool, changed func(bool)) {
	box := NewEUICheckbox(14, false, &value, label, changed, 1)
	box.GetWidget().LayoutData = widget.AnchorLayoutData{Padding: &widget.Insets{Left: x, Top: y + 7}}
	b.ui.Container.AddChild(box)
}

func hudColor(rgb uint32) color.NRGBA {
	return color.NRGBA{uint8(rgb >> 16), uint8(rgb >> 8), uint8(rgb), 255}
}

// Rasterize the bevel once. Nine-slicing preserves its rounded ends, inner
// lip and top light at every panel/button width without per-frame gradients.
func hudBevel(radius int, rimTop, rimBottom, faceTop, faceBottom uint32) *euiimage.NineSlice {
	const size = 40
	dc := gg.NewContext(size, size)
	layer := func(inset float64, top, bottom uint32) {
		g := gg.NewLinearGradient(0, inset, 0, size-inset)
		g.AddColorStop(0, hudColor(top))
		g.AddColorStop(1, hudColor(bottom))
		dc.SetFillStyle(g)
		dc.DrawRoundedRectangle(inset, inset, size-2*inset, size-2*inset, max(0, float64(radius)-inset))
		dc.Fill()
	}
	layer(0, 0x191918, 0x100f0e)
	layer(1, rimTop, rimBottom)
	layer(2, 0x302f2b, 0x5c5850)
	layer(3, faceTop, faceBottom)
	return euiimage.NewNineSliceBorder(ebiten.NewImageFromImage(dc.Image()), radius+1)
}

func hudButtons() (*widget.ButtonImage, *widget.ButtonImage) {
	normal := widget.ButtonImage{
		Idle:     hudBevel(13, 0x908e84, 0x514f4a, 0x575651, 0x393937),
		Hover:    hudBevel(13, 0xd8d0b8, 0x827e70, 0x66645c, 0x43423e),
		Pressed:  hudBevel(13, 0x716652, 0xb2a68c, 0x35332e, 0x514d43),
		Disabled: hudBevel(13, 0x575650, 0x42413c, 0x403f3c, 0x343432),
	}
	selected := normal
	selected.Idle = hudBevel(13, 0xd3b77a, 0x87724c, 0x665a42, 0x453d30)
	selected.Hover = hudBevel(13, 0xf1d694, 0xa58d5e, 0x7b6b4a, 0x544936)
	return &normal, &selected
}

// Use the same recessed slot and glossy overlay as the in-game hotbar.
// Draw the rim behind the artwork so the Buff's portrait aspect is unchanged.
func hudSlots() (*widget.ButtonImage, *widget.ButtonImage) {
	frame := func(rim uint32, overlay string) *euiimage.NineSlice {
		const size = 40
		slot := ebiten.NewImage(size, size)
		hudBevel(6, rim, 0x4c4941, 0x323232, 0x232321).Draw(slot, size, size, nil)
		ns := hotbarAtlasTexture.GetNineSlice(overlay)
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Scale(34/float64(ns.Width), 34/float64(ns.Height))
		op.GeoM.Translate(3, 3)
		slot.DrawImage(ns.Texture, op)
		return euiimage.NewNineSliceBorder(slot, 8)
	}
	normal := widget.ButtonImage{
		Idle:     frame(0x8c8980, "hotbar_empty.png"),
		Hover:    frame(0xeee3c8, "hotbar_fg.png"),
		Pressed:  frame(0xf5db99, "hotbar_clicked.png"),
		Disabled: frame(0x625e57, "hotbar_disabled.png"),
	}
	selected := normal
	selected.Idle = normal.Pressed
	selected.Hover = frame(0xffeec0, "hotbar_clicked.png")
	return &normal, &selected
}

func drawHUDWindow(dst *ebiten.Image, r image.Rectangle, skin *euiimage.NineSlice) {
	if r.Empty() {
		return
	}
	skin.Draw(dst, r.Dx(), r.Dy(), func(op *ebiten.DrawImageOptions) {
		op.Filter = ebiten.FilterLinear
		op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	})
}

func (b *StrategyBoard) drawWindows(screen *ebiten.Image) {
	w, h := b.width, b.height-bottomBar
	r := b.canvas()
	drawHUDWindow(screen, image.Rect(3, 2, w-3, topBar-2), hudTitle)
	drawHUDWindow(screen, image.Rect(3, h+2, w-3, b.height-2), hudWindow)
	if !b.compact {
		drawHUDWindow(screen, image.Rect(3, topBar, leftRail-2, h), hudWindow)
		drawHUDWindow(screen, image.Rect(w-rightRail+2, topBar, w-3, h), hudWindow)
	}
	if !b.compact || (b.panel != "工具" && b.panel != "对象") {
		drawHUDWindow(screen, image.Rect(r.Min.X+2, r.Max.Y+1, r.Max.X-2, h), hudWindow)
	}
	if b.compact {
		drawHUDWindow(screen, b.drawerRect(), hudWindow)
	}
}
