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
	Segments     []float64
	Interactable bool
	ClickAt      func(c float64, p float64)

	handler furex.ViewHandler
}

func (b *Bar) Handler() furex.ViewHandler {
	b.handler.Extra = b
	b.handler.Draw = b.Draw
	b.handler.JustReleasedMouseButtonLeft = b.HandleJustReleasedMouseButtonLeft
	b.handler.JustPressedMouseButtonLeft = b.HandleJustPressedMouseButtonLeft
	b.handler.MouseEnter = b.HandleMouseEnter
	b.handler.MouseLeave = b.HandleMouseLeave
	return b.handler
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
		return true
	}
	return false
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
	progress := 0.0
	switch b.Progress.(type) {
	case float64:
		progress = b.Progress.(float64)
	case func() float64:
		progress = b.Progress.(func() float64)()
	default:
		progress = 1.0
	}
	if progress > 1.0 {
		progress = 1.0
	}
	if progress < 0.0 {
		progress = 0.0
	}
	if len(b.Segments) == 0 {
		b.BG.Draw(screen, frame, nil)
	} else {
		s := 0.0
		for i := range b.Segments {
			e := b.Segments[i]
			b.BG.Draw(screen, image.Rect(
				frame.Min.X+int(s*float64(frame.Dx())),
				frame.Min.Y,
				frame.Min.X+int(e*float64(frame.Dx())),
				frame.Max.Y,
			), nil)
			s = e
		}
	}

	p := progress * float64(frame.Dx())
	frame.Max.X = frame.Min.X + int(p)
	b.FG.Draw(screen, frame, nil)
}
