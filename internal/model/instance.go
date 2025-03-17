package model

import (
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

type Instance struct {
	Status             *StatusData
	Face               float64
	Object             object.Object
	LastActive         int64
	casting            *Skill
	castHistory        []*Skill
	damageTakenHistory []DamageTaken
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

func (i *Instance) Render(parentStatus *StatusData, renderTargetRing bool, camera *CameraData, screen *ebiten.Image) {
	worldM := camera.WorldMatrixInverted()
	colorM := colorm.ColorM{}

	if parentStatus.IsDead() {
		colorM.ChangeHSV(0, 0, 1)
	}

	if renderTargetRing {
		geoM := texture.CenterGeoM(parentStatus.RingConfig.Texture)
		geoM.Rotate(i.Face)
		geoM.Scale(parentStatus.RingConfig.Scale, parentStatus.RingConfig.Scale)
		geoM.Translate(i.Object.Position()[0], i.Object.Position()[1])

		geoM.Concat(worldM)

		colorm.DrawImage(screen, parentStatus.RingConfig.Texture, colorM, &colorm.DrawImageOptions{
			GeoM: geoM,
		})
	}

	if parentStatus.RoleTexture() != nil {
		// role icon never rotate with player but camera
		geoM := texture.CenterGeoM(parentStatus.RoleTexture())
		geoM.Scale(0.5, 0.5)
		geoM.Rotate(camera.Rotation)
		geoM.Translate(i.Object.Position()[0], i.Object.Position()[1])
		geoM.Concat(worldM)

		colorm.DrawImage(screen, parentStatus.RoleTexture(), colorM, &colorm.DrawImageOptions{
			GeoM: geoM,
		})
	}

	if parentStatus.Charater != nil {
		geoM := texture.CenterGeoM(parentStatus.Charater)
		geoM.Translate(i.Object.Position()[0], i.Object.Position()[1])
		geoM.Concat(worldM)

		colorm.DrawImage(screen, parentStatus.Charater, colorM, &colorm.DrawImageOptions{
			GeoM: geoM,
		})
	}
}
