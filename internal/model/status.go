package model

import (
	"fmt"

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
	death    bool
}

func (r *StatusData) Reset() {
	r.HP = r.MaxHP
	r.Mana = r.MaxMana
	r.death = false
	r.BuffList.Clear()
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
	if r.Role == role.Special {
		return texture.NewTextureFromFile(fmt.Sprintf("asset/boss/%d.png", r.GameID))
	}

	return texture.NewTextureFromFile("asset/role/" + r.Role.String() + ".png")
}
