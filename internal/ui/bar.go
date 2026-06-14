package ui

import (
	"image"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

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
