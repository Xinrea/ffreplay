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
}

func (s *Sprite) Draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View) {
	if s.Texture != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(frame.Dx())/float64(s.Texture.Bounds().Dx()), float64(frame.Dy())/float64(s.Texture.Bounds().Dy()))
		op.GeoM.Translate(float64(frame.Min.X), float64(frame.Min.Y))
		screen.DrawImage(s.Texture, op)
		return
	}
	s.NineSliceTexture.Draw(screen, frame)
}
