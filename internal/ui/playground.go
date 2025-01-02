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
	return &PlaygroundUI{
		view: &furex.View{},
	}
}

func (p *PlaygroundUI) Update(w, h int) {
	s := ebiten.Monitor().DeviceScaleFactor()
	furex.GlobalScale = s
	p.view.UpdateWithSize(w, h)
}

func (p *PlaygroundUI) Draw(screen *ebiten.Image) {

}
