package ui

import (
	"image"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

type Sprite struct {
	NineSliceTexture *texture.NineSlice
	Texture          *ebiten.Image
	BlendAlpha       bool
	Alpha            float32
}

func (s *Sprite) Draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View) {
	if s.Texture != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(frame.Dx())/float64(s.Texture.Bounds().Dx()), float64(frame.Dy())/float64(s.Texture.Bounds().Dy()))
		op.GeoM.Translate(float64(frame.Min.X), float64(frame.Min.Y))
		if s.BlendAlpha {
			op.ColorScale.ScaleAlpha(s.Alpha)
		}
		screen.DrawImage(s.Texture, op)
		return
	}
	var scale *ebiten.ColorScale = nil
	if s.BlendAlpha {
		scale = &ebiten.ColorScale{}
		scale.ScaleAlpha(s.Alpha)
	}
	s.NineSliceTexture.Draw(screen, frame, scale)
}
