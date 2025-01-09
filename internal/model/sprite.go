package model

import (
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteData is a struct that contains texture, face, and object.
// Face is the radian relative to the north direction, range: [-pi, pi].
type SpriteData struct {
	Texture     *ebiten.Image
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
		geoM := texture.CenterGeoM(s.Texture)
		geoM.Rotate(instance.Face)
		geoM.Translate(instance.Object.Position()[0], instance.Object.Position()[1])

		wordM := camera.WorldMatrixInverted()
		geoM.Concat(wordM)

		screen.DrawImage(s.Texture, &ebiten.DrawImageOptions{
			GeoM: geoM,
		})
	}
}

func (i *Instance) IsActive(tick int64) bool {
	if i.LastActive == -1 {
		return false
	}

	if i.casting != nil {
		i.LastActive = tick

		return true
	}

	return util.TickToMS(tick-i.LastActive) <= 2500
}

func (i *Instance) Cast(gameSkill Skill) {
	// just auto attack
	if gameSkill.ID == 7 || gameSkill.ID == 8 {
		return
	}

	// maybe the skill to cast is the effect of previous long casting skill
	if isSucceed(i.casting, &gameSkill) {
		return
	}

	if len(i.historyCasting) > 0 && isSucceed(i.historyCasting[len(i.historyCasting)-1], &gameSkill) {
		return
	}
	// no need to spell, just move into historyCasting
	if gameSkill.Cast == 0 {
		if i.casting != nil && i.casting.Cast > 0 {
			i.ClearCast()
		}

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
	if i.casting == nil {
		return
	}

	i.historyCasting = append(i.historyCasting, i.casting)
	i.casting = nil
}

func (i *Instance) GetHistoryCast(tick int64) []*Skill {
	// only keep last 5 gcd of history
	cnt := 5

	for k := len(i.historyCasting) - 1; k >= 0; k-- {
		if i.historyCasting[k].IsGCD {
			cnt--
		}

		if cnt == 0 {
			i.historyCasting = i.historyCasting[k:]

			break
		}
	}

	return i.historyCasting
}

func (i *Instance) Reset() {
	i.LastActive = -1
	i.casting = nil
	i.historyCasting = nil
}

// isSucceed checks same skill that cast twice, but previous one has cast time.
func isSucceed(x, y *Skill) bool {
	if x == nil || y == nil {
		return false
	}

	if x.ID != y.ID {
		return false
	}

	return x.Cast > 0 && y.Cast == 0
}
