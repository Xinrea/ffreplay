package model

import (
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/yohamta/donburi"
)

type RoleType string

const (
	MT          RoleType = "mt"
	ST          RoleType = "st"
	H1          RoleType = "h1"
	H2          RoleType = "h2"
	D1          RoleType = "d1"
	D2          RoleType = "d2"
	D3          RoleType = "d3"
	D4          RoleType = "d4"
	Paladin     RoleType = "Paladin"
	Warrior     RoleType = "Warrior"
	DarkKnight  RoleType = "DarkKnight"
	Gunbreaker  RoleType = "Gunbreaker"
	WhiteMage   RoleType = "WhiteMage"
	Scholar     RoleType = "Scholar"
	Astrologian RoleType = "Astrologian"
	Sage        RoleType = "Sage"
	Monk        RoleType = "Monk"
	Dragoon     RoleType = "Dragoon"
	Ninja       RoleType = "Ninja"
	Samurai     RoleType = "Samurai"
	Reaper      RoleType = "Reaper"
	Viper       RoleType = "Viper"
	Bard        RoleType = "Bard"
	Machinist   RoleType = "Machinist"
	Dancer      RoleType = "Dancer"
	BlackMage   RoleType = "BlackMage"
	Summoner    RoleType = "Summoner"
	RedMage     RoleType = "RedMage"
	Pictomancer RoleType = "Pictomancer"
	Boss        RoleType = "Boss"
	NPC         RoleType = "NPC"
)

var Status = donburi.NewComponentType[StatusData]()

type StatusData struct {
	GameID     int64
	ID         int64
	Name       string
	Role       RoleType
	HP         int
	MaxHP      int
	Mana       int
	MaxMana    int
	IsBoss     bool
	BuffList   *BuffList
	Casting    *Cast
	LastActive int64
}

type Ability struct {
	Guid        int64
	Name        string
	AbilityIcon string
}

type Cast struct {
	ApplyTick int64
	Ability   Ability
	Duration  int64
}

func (r *StatusData) Reset() {
	r.HP = r.MaxHP
	r.Mana = r.MaxMana
	r.BuffList = NewBuffList()
}

func (r *StatusData) TakeDamage(d Damage) {
	r.BuffList.ProcessDamage(&d)
	r.HP -= d.Amount
	if r.HP <= 0 {
		r.HP = 0
	}
}

func (r *StatusData) IsDead() bool {
	return r.HP <= 0
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

func (r StatusData) RoleTexture() *texture.Texture {
	return texture.NewTextureFromFile("asset/role/" + string(r.Role) + ".png")
}
