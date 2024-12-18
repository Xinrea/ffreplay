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
	Scale       float64
	Instances   []*Instance
	Initialized bool
}

type Instance struct {
	Face       float64
	Object     object.Object
	LastActive int64
	Casting    *Skill
}

func (s SpriteData) Render(camera *CameraData, screen *ebiten.Image) {
	if !s.Initialized {
		return
	}
	for _, instance := range s.Instances {
		geoM := s.Texture.GetGeoM()
		geoM.Rotate(instance.Face)
		geoM.Translate(instance.Object.Position()[0], instance.Object.Position()[1])
		wordM := camera.WorldMatrix()
		wordM.Invert()
		geoM.Concat(wordM)
		screen.DrawImage(s.Texture.Img(), &ebiten.DrawImageOptions{
			GeoM: geoM,
		})
	}
}

func (i *Instance) Cast(gameSkill Skill) {
	i.Casting = &gameSkill
}
