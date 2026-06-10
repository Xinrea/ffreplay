package ui

import (
	"image"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/ebitenui/ebitenui/widget"
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
			if p, ok := b.Progress.(float64); ok {
				progress = p
			}
		case func() float64:
			if p, ok := b.Progress.(func() float64); ok {
				progress = p()
			}
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
	drawNineSliceBar(screen, frame, b.BG, b.FG, resolveBarProgress(b.Progress), b.Segments)
}

func resolveBarProgress(progress any) float64 {
	p := 1.0
	switch v := progress.(type) {
	case float64:
		p = v
	case func() float64:
		p = v()
	}
	if p > 1.0 {
		return 1.0
	}
	if p < 0.0 {
		return 0.0
	}
	return p
}

func drawNineSliceBar(
	screen *ebiten.Image,
	frame image.Rectangle,
	bg *texture.NineSlice,
	fg *texture.NineSlice,
	progress float64,
	segments []float64,
) {
	if len(segments) == 0 {
		bg.Draw(screen, frame, nil)
	} else {
		start := 0.0
		for _, end := range segments {
			bg.Draw(screen, image.Rect(
				frame.Min.X+int(start*float64(frame.Dx())),
				frame.Min.Y,
				frame.Min.X+int(end*float64(frame.Dx())),
				frame.Max.Y,
			), nil)
			start = end
		}
	}

	fgFrame := frame
	fgFrame.Max.X = fgFrame.Min.X + int(progress*float64(frame.Dx()))
	fg.Draw(screen, fgFrame, nil)
}

type euiBar struct {
	widget   *widget.Widget
	width    int
	height   int
	bg       *texture.NineSlice
	fg       *texture.NineSlice
	progress any
	segments []float64
}

func NewEUIBar(width, height int, bg *texture.NineSlice, fg *texture.NineSlice, progress any, segments []float64, layoutData any) *euiBar {
	bar := &euiBar{
		width:    width,
		height:   height,
		bg:       bg,
		fg:       fg,
		progress: progress,
		segments: segments,
	}
	opts := []widget.WidgetOpt{
		widget.WidgetOpts.LayoutData(layoutData),
	}
	bar.widget = widget.NewWidget(opts...)
	return bar
}

func (b *euiBar) GetWidget() *widget.Widget {
	return b.widget
}

func (b *euiBar) SetLocation(rect image.Rectangle) {
	b.widget.Rect = rect
}

func (b *euiBar) PreferredSize() (int, int) {
	return b.width, b.height
}

func (b *euiBar) Validate() {}

func (b *euiBar) Update(updObj *widget.UpdateObject) {
	b.widget.Update(updObj)
}

func (b *euiBar) Render(screen *ebiten.Image) {
	b.widget.Render(screen)
	drawNineSliceBar(screen, b.widget.Rect, b.bg, b.fg, resolveBarProgress(b.progress), b.segments)
}
