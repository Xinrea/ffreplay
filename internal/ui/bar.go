package ui

import (
	"image"
	"log"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

type Bar struct {
	Progress any
	BG       *texture.NineSlice
	FG       *texture.NineSlice
	ClickAt  func(p float64)
}

func (b *Bar) HandleJustPressedMouseButtonLeft(x, y int) bool {
	log.Println("Bar.HandleJustPressedMouseButtonLeft", x, y)
	return true
}

func (b *Bar) HandleJustReleasedMouseButtonLeft(x, y int) {
}

func (b *Bar) Draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View) {
	b.BG.Draw(screen, frame)
	progress := 0.0
	switch b.Progress.(type) {
	case float64:
		progress = b.Progress.(float64)
	case func() float64:
		progress = b.Progress.(func() float64)()
	default:
		return
	}
	p := progress * float64(frame.Dx())
	frame.Max.X = frame.Min.X + int(p)
	b.FG.Draw(screen, frame)
}
