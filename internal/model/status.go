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
	GameID      int64
	ID          int64
	Name        string
	Role        role.RoleType
	HP          int
	MaxHP       int
	Mana        int
	MaxMana     int
	BuffList    *BuffList
	Charater    *ebiten.Image
	Marker      int
	HeadMarkers []*HeadMarker
	Tethers     []Tether
	Instances   []*Instance
	RingConfig  RingConfiguration
	death       bool
}

const STANDARD_RING_SCALE = 0.5

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
		if !renderAllInstances {
			if (r.Role == role.NPC || r.Role == role.Boss) && !instance.IsActive(tick) {
				continue
			}
		}

		instance.Render(r, renderTargetRing, camera, screen)

		// render head markers, marker position is relative to the instance position
		colorM := colorm.ColorM{}

		for _, headMarker := range r.HeadMarkers {
			pos := instance.Object.Position()
			geoM := texture.CenterGeoM(headMarker.Texture())
			geoM.Scale(r.RingConfig.Scale/STANDARD_RING_SCALE, r.RingConfig.Scale/STANDARD_RING_SCALE)
			geoM.Rotate(camera.Rotation)
			geoM.Translate(pos[0], pos[1])
			geoM.Concat(wordM)

			colorm.DrawImage(screen, headMarker.Texture(), colorM, &colorm.DrawImageOptions{
				GeoM: geoM,
			})
		}
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

func (r *StatusData) AddHeadMarker(t HeadMarkerType) {
	r.HeadMarkers = append(r.HeadMarkers, NewHeadMarker(t))
}

func (r *StatusData) ClearHeadMarker() {
	r.HeadMarkers = nil
}
