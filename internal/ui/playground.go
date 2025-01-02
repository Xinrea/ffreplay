package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

type PlaygroundUI struct {
	view *furex.View
}

var _ UI = (*PlaygroundUI)(nil)

func NewPlaygroundUI() *PlaygroundUI {
	checked := true
	view := &furex.View{
		Position: furex.PositionAbsolute,
		Top:      0,
		Left:     0,
	}
	view.AddChild(MultiCheckBoxView(32, &checked))
	return &PlaygroundUI{
		view: view,
	}
}

func (p *PlaygroundUI) Update(w, h int) {
	s := ebiten.Monitor().DeviceScaleFactor()
	furex.GlobalScale = s
	p.view.UpdateWithSize(w, h)
}

func (p *PlaygroundUI) Draw(screen *ebiten.Image) {
	p.view.Draw(screen)
}
