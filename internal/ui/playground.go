package ui

import (
	"image"
	"sync"

	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
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
			command.Attrs.MarginBottom = 20
			command.Attrs.MarginLeft = 20
			partyList := NewPartyList(nil)
			partyList.Attrs.MarginTop = 40
			partyList.Attrs.MarginLeft = 20
			p.base.AddChild(partyList)
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
