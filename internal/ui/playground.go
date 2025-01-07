package ui

import (
	"image"
	"sync"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
	"golang.org/x/image/math/f64"
)

// Due to the design of furex.View, root view cannot handle any events,
// so we need a global root as wrap.
var root = furex.NewView(furex.ID("Root"))

type PlaygroundUI struct {
	base *furex.View
	once sync.Once
}

var _ UI = (*PlaygroundUI)(nil)

func NewPlaygroundUI(ecs *ecs.ECS) *PlaygroundUI {
	ecsInstance = ecs
	baseWrap := furex.NewView(
		furex.ID("Playground"),
		furex.Direction(furex.Column),
		furex.Justify(furex.JustifySpaceBetween),
		furex.Grow(1),
	)
	baseWrap.Handler.JustPressedMouseButtonLeft = func(frame image.Rectangle, x int, y int) bool {
		for _, c := range baseWrap.FilterByTagName("input") {
			if fh, ok := c.Handler.Extra.(Focusable); ok {
				fh.SetFocus(false)
			}
		}

		entry.GetGlobal(ecsInstance).UIFocus = false

		return false
	}

	root.AddChild(baseWrap)

	return &PlaygroundUI{
		base: baseWrap,
	}
}

func (p *PlaygroundUI) Update(w, h int) {
	global := entry.GetGlobal(ecsInstance)
	if global.Loaded.Load() {
		p.once.Do(func() {
			command := CommandView()
			command.Attrs.MarginBottom = UIPadding
			command.Attrs.MarginLeft = UIPadding

			hotbar := HotBarView(2, 8)
			hotbar.Attrs.MarginTop = UIPadding
			hotbar.Attrs.MarginRight = UIPadding

			p.SetupHotBar(hotbar, 2, 8)

			topView := furex.NewView(
				furex.Grow(1),
				furex.Direction(furex.Row),
				furex.Justify(furex.JustifySpaceBetween),
			)

			partyList := NewPartyList(nil)
			partyList.Attrs.MarginTop = 40
			partyList.Attrs.MarginLeft = UIPadding

			topView.AddChild(partyList)
			topView.AddChild(hotbar)

			p.base.AddChild(topView)
			p.base.AddChild(command)
		})
	}

	s := ebiten.Monitor().DeviceScaleFactor()
	furex.GlobalScale = s

	root.UpdateWithSize(w, h)
}

func (p *PlaygroundUI) Draw(screen *ebiten.Image) {
	root.Draw(screen)
}

func (p *PlaygroundUI) SetupHotBar(v *furex.View, w, h int) {
	newWorldMarkerHotBarItem := func(marker model.WorldMarkerType) *furex.View {
		return HotbarItemView(&HotBarItemConfig{
			Name: "test",
			Icon: model.WorldMarkerConfigs[marker].Texture,
			ClickHandler: func() {
				global := entry.GetGlobal(ecsInstance)
				camera := entry.GetCamera(ecsInstance)
				// if marker exists, remove it
				for markerEntry := range component.WorldMarker.Iter(ecsInstance.World) {
					markerData := component.WorldMarker.Get(markerEntry)
					if markerData.Type == marker {
						markerEntry.Remove()

						return
					}
				}

				x, y := ebiten.CursorPosition()
				wx, wy := camera.ScreenToWorld(float64(x), float64(y))
				entry.NewWorldMarker(ecsInstance, marker, f64.Vec2{wx, wy})
				global.WorldMarkerSelected = int(marker)
			},
		})
	}

	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			marker := model.WorldMarkerType(i*w + j)
			if marker > model.WorldMarker4 {
				return
			}

			v.NthChild(i).NthChild(j).ReplaceWith(newWorldMarkerHotBarItem(marker))
		}
	}
}
