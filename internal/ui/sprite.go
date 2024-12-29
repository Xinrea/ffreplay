package ui

import (
	"image"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

type Sprite struct {
	Texture *texture.NineSlice
}

func (s *Sprite) Draw(screen *ebiten.Image, frame image.Rectangle, view *furex.View) {
	s.Texture.Draw(screen, frame)
}
