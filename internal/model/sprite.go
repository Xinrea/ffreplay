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
	Face               float64
	Object             object.Object
	LastActive         int64
	casting            *Skill
	castHistory        []*Skill
	damageTakenHistory []DamageTaken
	tethers            []*Instance
}

type DamageTaken struct {
	Tick         int64
	Type         DamageType
	SourceID     int64
	Ability      Skill
	Amount       int64
	Multiplier   float64
	RelatedBuffs []*BasicBuffInfo
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

func (i *Instance) AddTether(tether *Instance) {
	i.tethers = append(i.tethers, tether)
}

func (i *Instance) GetTethers() []*Instance {
	return i.tethers
}

func (i *Instance) ClearTether() {
	i.tethers = nil
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

func (i *Instance) Cast(gameSkill *Skill) {
	// just auto attack
	if gameSkill.ID == 7 || gameSkill.ID == 8 {
		return
	}

	// maybe the skill to cast is the effect of previous long casting skill
	if isSucceed(i.casting, gameSkill) {
		return
	}

	if len(i.castHistory) > 0 && isSucceed(i.castHistory[len(i.castHistory)-1], gameSkill) {
		return
	}
	// no need to spell, just move into historyCasting
	if gameSkill.Cast == 0 {
		if i.casting != nil && i.casting.Cast > 0 {
			i.DoneCast()
		}

		i.castHistory = append(i.castHistory, gameSkill)

		return
	}

	i.DoneCast()

	i.casting = gameSkill
}

func (i *Instance) GetCast() *Skill {
	return i.casting
}

func (i *Instance) DoneCast() {
	if i.casting == nil {
		return
	}

	i.castHistory = append(i.castHistory, i.casting)
	i.casting = nil
}

func (i *Instance) GetHistoryCast(tick int64) []*Skill {
	// only keep last 5 gcd of history
	cnt := 5

	for k := len(i.castHistory) - 1; k >= 0; k-- {
		if i.castHistory[k].IsGCD {
			cnt--
		}

		if cnt == 0 {
			i.castHistory = i.castHistory[k:]

			break
		}
	}

	return i.castHistory
}

func (i *Instance) AddDamageTaken(damage DamageTaken) {
	// TODO: ignore dots, bleeding, etc for now.(type:1, 64)
	if damage.Type != Physical && damage.Type != Magical && damage.Type != Special {
		return
	}

	i.damageTakenHistory = append(i.damageTakenHistory, damage)
}

// GetHistoryDamageTaken returns the last n damage taken history.
func (i *Instance) GetHistoryDamageTaken(n int) []DamageTaken {
	if len(i.damageTakenHistory) > n {
		return i.damageTakenHistory[len(i.damageTakenHistory)-n:]
	}

	return i.damageTakenHistory
}

func (i *Instance) Reset() {
	i.LastActive = -1
	i.casting = nil
	i.castHistory = nil
	i.damageTakenHistory = nil
	i.tethers = nil
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
