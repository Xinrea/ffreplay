package model

import (
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/yohamta/donburi"
)

var Status = donburi.NewComponentType[StatusData]()

type StatusData struct {
	GameID     int64
	ID         int64
	Name       string
	Role       role.RoleType
	HP         int
	MaxHP      int
	Mana       int
	MaxMana    int
	BuffList   *BuffList
	Marker     int
	Tethers    []Tether
	Instances  []*Instance
	RingConfig RingConfiguration
	death      bool
}

type RingConfiguration struct {
	Texture *ebiten.Image
	Scale   float64
}

type Tether struct {
	ApplyTick int64
	Target    *StatusData
}

func (r *StatusData) Render(
	tick int64,
	camera *CameraData,
	screen *ebiten.Image,
	renderTargetRing bool,
	renderAllInstances bool,
) {
	if !renderAllInstances && r.Role == role.NPC {
		return
	}

	wordM := camera.WorldMatrixInverted()

	for _, instance := range r.Instances {
		if !renderAllInstances && !instance.IsActive(tick) {
			continue
		}

		colorM := colorm.ColorM{}

		if r.IsDead() {
			colorM.ChangeHSV(0, 0, 1)
		}

		if renderTargetRing {
			geoM := texture.CenterGeoM(r.RingConfig.Texture)
			geoM.Rotate(instance.Face)
			geoM.Scale(r.RingConfig.Scale, r.RingConfig.Scale)
			geoM.Translate(instance.Object.Position()[0], instance.Object.Position()[1])

			geoM.Concat(wordM)

			colorm.DrawImage(screen, r.RingConfig.Texture, colorM, &colorm.DrawImageOptions{
				GeoM: geoM,
			})
		}

		if r.RoleTexture() == nil {
			continue
		}

		geoM := texture.CenterGeoM(r.RoleTexture())
		geoM.Scale(0.5, 0.5)
		geoM.Translate(instance.Object.Position()[0], instance.Object.Position()[1])
		geoM.Concat(wordM)

		colorm.DrawImage(screen, r.RoleTexture(), colorM, &colorm.DrawImageOptions{
			GeoM: geoM,
		})
	}
}

func (r *StatusData) Reset() {
	r.HP = r.MaxHP
	r.Mana = r.MaxMana
	r.death = false
	r.BuffList.Clear()
	r.ClearTether()
}

func (r *StatusData) TakeDamage(d Damage) {
	r.BuffList.ProcessDamage(&d)
	r.HP -= d.Amount

	if r.HP <= 0 {
		r.HP = 0
	}
}

func (r *StatusData) SetDeath(b bool) {
	r.death = b
}

func (r *StatusData) IsDead() bool {
	return r.death
}

func (r *StatusData) TakeHeal(h Heal) {
	r.BuffList.ProcessHeal(&h)

	if r.IsDead() {
		return
	}

	r.HP += h.Amount

	if r.HP > r.MaxHP {
		r.HP = r.MaxHP
	}
}

func (r StatusData) RoleTexture() *ebiten.Image {
	if r.Role == role.Boss {
		return nil
	}

	return texture.NewTextureFromFile("asset/role/" + r.Role.String() + ".png")
}

func (r *StatusData) AddTether(tick int64, target *StatusData) {
	r.Tethers = append(r.Tethers, Tether{
		ApplyTick: tick,
		Target:    target,
	})
}

func (r *StatusData) GetTethers() []Tether {
	return r.Tethers
}

func (r *StatusData) ClearTether() {
	r.Tethers = nil
}
