package model

import (
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

var Status = donburi.NewComponentType[StatusData]()

type StatusData struct {
	GameID   int64
	ID       int64
	Name     string
	Role     role.RoleType
	HP       int
	MaxHP    int
	Mana     int
	MaxMana  int
	BuffList *BuffList
	Marker   int
	Tethers  []Tether
	death    bool
}

type Tether struct {
	ApplyTick int64
	Target    *SpriteData
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
	return texture.NewTextureFromFile("asset/role/" + r.Role.String() + ".png")
}

func (r *StatusData) AddTether(tick int64, target *SpriteData) {
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
