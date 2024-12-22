package model

import (
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
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
	Face           float64
	Object         object.Object
	LastActive     int64
	casting        *Skill
	historyCasting []*Skill
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

func (i *Instance) IsActive(tick int64) bool {
	if i.LastActive == -1 {
		return false
	}
	return util.TickToMS(tick-i.LastActive) <= 2500
}

func (i *Instance) Cast(gameSkill Skill) {
	// maybe the skill to cast is the effect of previous long casting skill
	if i.casting != nil && i.casting.ID == gameSkill.ID && i.casting.Cast > 0 {
		return
	}
	// no need to spell, just move into historyCasting
	if gameSkill.Cast == 0 {
		i.historyCasting = append(i.historyCasting, &gameSkill)
		return
	}
	if i.casting != nil {
		i.ClearCast()
	}
	i.casting = &gameSkill
}

func (i *Instance) GetCast() *Skill {
	return i.casting
}

func (i *Instance) ClearCast() {
	i.historyCasting = append(i.historyCasting, i.casting)
	i.casting = nil
}

func (i *Instance) GetHistoryCast(tick int64) []*Skill {
	ret := []*Skill{}
	for _, c := range i.historyCasting {
		if tick-c.StartTick > 10*60 /*10s*/ {
			continue
		}
		ret = append(ret, c)
	}
	i.historyCasting = ret
	return ret
}

func (i *Instance) Reset() {
	i.LastActive = -1
	i.casting = nil
	i.historyCasting = nil
}
