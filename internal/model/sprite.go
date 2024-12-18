package model

import (
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteData is a struct that contains texture, face, and object
// Face is the radian relative to the north direction, range: [-pi, pi]
type SpriteData struct {
	Texture     *texture.Texture
	Face        float64
	Scale       float64
	Object      object.Object
	Initialized bool
}

func (s SpriteData) Render(camera *CameraData, screen *ebiten.Image) {
	if !s.Initialized {
		return
	}
	geoM := s.Texture.GetGeoM()
	geoM.Rotate(s.Face)
	geoM.Translate(s.Object.Position()[0], s.Object.Position()[1])
	wordM := camera.WorldMatrix()
	wordM.Invert()
	geoM.Concat(wordM)
	screen.DrawImage(s.Texture.Img(), &ebiten.DrawImageOptions{
		GeoM: geoM,
	})
}
