package ui

import (
	"image"
	"time"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

var hotbarAtlasTexture = texture.NewTextureAtlasFromFile("asset/ui/hotbar.xml")

func HotBarView(w, h int) *furex.View {
	view := furex.NewView(
		furex.TagName("hotbar"),
		furex.Direction(furex.Column),
	)

	for i := 0; i < h; i++ {
		row := furex.NewView(
			furex.TagName("hotbar-row"),
			furex.MarginTop(2),
		)

		for j := 0; j < w; j++ {
			item := &HotBarItemConfig{}
			row.AddChild(HotbarItemView(item))
		}

		view.AddChild(row)
	}

	return view
}

type HotBarItemConfig struct {
	Name         string
	Icon         *ebiten.Image
	ClickHandler func()

	clickTime time.Time
}

var globalHoveredHotbarItem *furex.View

func HotbarItemView(item *HotBarItemConfig) *furex.View {
	view := furex.NewView(
		furex.TagName("hotbar-item"),
		furex.Width(48),
		furex.Height(48),
		furex.MarginLeft(2),
	)

	view.Handler.MouseEnter = func(x, y int) bool {
		globalHoveredHotbarItem = view

		ebiten.SetCursorShape(ebiten.CursorShapePointer)

		return true
	}

	view.Handler.MouseLeave = func() {
		if globalHoveredHotbarItem == nil || globalHoveredHotbarItem == view {
			ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		}

		globalHoveredHotbarItem = nil
	}

	if item.Icon == nil {
		view.AddChild(furex.NewView(
			furex.TagName("hotbar-empty"),
			furex.Position(furex.PositionAbsolute),
			furex.Top(0),
			furex.Left(0),
			furex.Width(48),
			furex.Height(48),
			furex.Handler(
				&Sprite{
					NineSliceTexture: hotbarAtlasTexture.GetNineSlice("hotbar_empty.png"),
				},
			)))

		return view
	}

	view.AddChild(furex.NewView(
		furex.TagName("hotbar-icon"),
		furex.Width(48),
		furex.Height(48),
		furex.Handler(
			&Sprite{
				Texture: item.Icon,
			},
		)))

	view.AddChild(furex.NewView(
		furex.TagName("hotbar-fg"),
		furex.Position(furex.PositionAbsolute),
		furex.Top(0),
		furex.Left(0),
		furex.Width(48),
		furex.Height(48),
		furex.Handler(
			&Sprite{
				NineSliceTexture: hotbarAtlasTexture.GetNineSlice("hotbar_fg.png"),
			},
		)))

	view.AddChild(furex.NewView(
		furex.TagName("hotbar-click"),
		furex.Hidden(true),
		furex.Position(furex.PositionAbsolute),
		furex.Top(0),
		furex.Left(0),
		furex.Width(48),
		furex.Height(48),
		furex.Handler(
			&Sprite{
				NineSliceTexture: hotbarAtlasTexture.GetNineSlice("hotbar_clicked.png"),
			},
		)))

	view.Handler.JustPressedMouseButtonLeft = func(frame image.Rectangle, x, y int) bool {
		item.clickTime = time.Now()

		if item.ClickHandler != nil {
			item.ClickHandler()
		}

		return true
	}

	view.Handler.Update = func(v *furex.View) {
		if time.Since(item.clickTime) < time.Millisecond*100 {
			view.FilterByTagName("hotbar-click")[0].Attrs.Hidden = false
		} else {
			view.FilterByTagName("hotbar-click")[0].Attrs.Hidden = true
		}
	}

	return view
}
