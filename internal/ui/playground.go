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
	view := &furex.View{
		Position: furex.PositionAbsolute,
		Top:      0,
		Left:     0,
	}
	view.AddChild(CheckBoxView(16, false, false, "test", nil))
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
