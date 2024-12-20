package model

import (
	"sort"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type BuffType int

const (
	NormalBuff BuffType = iota
	Debuff
)

type BuffList struct {
	buffs []Buff
}

func NewBuffList() *BuffList {
	return &BuffList{
		buffs: make([]Buff, 0, 24),
	}
}

func (bl *BuffList) SetBuffs(buffs []Buff) {
	bl.buffs = append(bl.buffs, buffs...)
}

func (bl *BuffList) Buffs() []Buff {
	// order buffs by type, debuff first
	sort.Slice(bl.buffs, func(i, j int) bool {
		return bl.buffs[i].Type > bl.buffs[j].Type
	})
	return bl.buffs
}

func (bl *BuffList) DeBuffs() (ret []Buff) {
	for _, b := range bl.buffs {
		if b.Type == Debuff {
			ret = append(ret, b)
		}
	}
	return
}

func (bl *BuffList) Add(buff Buff) {
	index := -1
	for i, b := range bl.buffs {
		if b.ID == buff.ID {
			index = i
			break
		}
	}
	if index != -1 {
		bl.buffs[index] = buff
	} else {
		bl.buffs = append(bl.buffs, buff)
	}
}

func (bl *BuffList) Refresh(buff Buff) {
	index := -1
	for i, b := range bl.buffs {
		if b.ID == buff.ID {
			index = i
			break
		}
	}
	if index != -1 {
		bl.buffs[index].Duration = buff.Duration
	}
}

func (bl *BuffList) Remove(buff Buff) {
	index := -1
	for i, b := range bl.buffs {
		if b.ID == buff.ID {
			index = i
			break
		}
	}
	if index != -1 {
		bl.buffs[index].Remove()
		bl.buffs = append(bl.buffs[:index], bl.buffs[index+1:]...)
	}
}

func (bl *BuffList) UpdateExpire(now int64) {
	toRemove := make([]Buff, 0)
	for _, b := range bl.buffs {
		if b.Expired(now) {
			toRemove = append(toRemove, b)
		}
	}
	for _, b := range toRemove {
		bl.Remove(b)
	}
}

func (bl *BuffList) ProcessDamage(damage *Damage) {
	for _, b := range bl.buffs {
		if b.ProcessDamage != nil {
			b.ProcessDamage(damage)
		}
	}
}

func (bl *BuffList) ProcessHeal(heal *Heal) {
	for _, b := range bl.buffs {
		if b.ProcessHeal != nil {
			b.ProcessHeal(heal)
		}
	}
}

func (bl *BuffList) Clear() {
	bl.buffs = make([]Buff, 0)
}

type Buff struct {
	Type   BuffType
	ID     int64
	Name   string
	Icon   string
	Stacks int
	// buff duration in ms
	Duration       int64
	ApplyTick      int64
	ECS            *ecs.ECS                                              `json:"-"`
	Source         *donburi.Entry                                        `json:"-"`
	Target         *donburi.Entry                                        `json:"-"`
	ProcessDamage  func(damage *Damage)                                  `json:"-"`
	ProcessHeal    func(heal *Heal)                                      `json:"-"`
	RemoveCallback func(*Buff, *ecs.ECS, *donburi.Entry, *donburi.Entry) `json:"-"`
}

func (b Buff) Remain(now int64) int64 {
	if b.Duration == 0 || b.Duration > 7200*1000 {
		return 0
	}
	return (b.Duration - util.TickToMS(now-b.ApplyTick)) / 1000
}

func (b Buff) Expired(now int64) bool {
	if b.Duration == 0 {
		return false
	}
	return b.Duration < util.TickToMS(now-b.ApplyTick)
}

func (b Buff) Texture() *texture.Texture {
	return texture.NewAbilityTexture(b.Icon)
}

func (b *Buff) Remove() {
	if b.RemoveCallback != nil {
		b.RemoveCallback(b, b.ECS, b.Source, b.Target)
	}
}
