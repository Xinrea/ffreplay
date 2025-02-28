package model

import (
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteData is a struct that contains texture and position data for rendering sprites.
type SpriteData struct {
	Texture *ebiten.Image
	Scale   float64
	Radian  float64
	Pos     vector.Vector
}
