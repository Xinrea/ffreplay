package ui

import (
	"image"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

type Bar struct {
	Progress     any
	BG           *texture.NineSlice
	FG           *texture.NineSlice
	Interactable bool
	ClickAt      func(c float64, p float64)
}

func (b *Bar) HandleJustPressedMouseButtonLeft(frame image.Rectangle, x, y int) bool {
	if b.ClickAt != nil {
		progress := 0.0
		switch b.Progress.(type) {
		case float64:
			progress = b.Progress.(float64)
		case func() float64:
			progress = b.Progress.(func() float64)()
		}
		b.ClickAt(progress, float64(x-frame.Min.X)/float64(frame.Dx()))
	}
	return true
}

func (b *Bar) HandleJustReleasedMouseButtonLeft(frame image.Rectangle, x, y int) {
}

func (b *Bar) HandleMouseEnter(x, y int) bool {
	if b.Interactable {
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	}
	return true
}

func (b *Bar) HandleMouseLeave() {
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)
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
