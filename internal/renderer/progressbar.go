package renderer

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type ProgressBar struct {
	w     int
	h     int
	Color color.Color
}

func NewProgressBar(w, h int, color color.Color) *ProgressBar {
	return &ProgressBar{
		w:     w,
		h:     h,
		Color: color,
	}
}

func (p *ProgressBar) Render(canvas *ebiten.Image, x float64, y float64, progress float64) {
	progressWidth := float64(p.w) * progress
	DrawFilledRect(canvas, x, y, progressWidth, float64(p.h), p.Color)
	StrokeRect(canvas, x, y, float64(p.w), float64(p.h), 1, p.Color)
}
