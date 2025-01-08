package model

import (
	"sort"

	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type BuffType int

const (
	NormalBuff BuffType = iota
	Debuff
)

// BuffFilter contains all useless buff id:
// https://github.com/Xinrea/ffreplay/issues/17
var BuffFilter = map[int64]bool{}

func initBuffFilter() {
	ids := []int64{
		1061, 1079, 1080, 1081, 1082, 1083, 1084, 1085, 1086,
		353, 354, 355, 356, 357, 360, 361, 362, 363, 364, 365, 366, 367, 368,
		413, 414, 902, 2932,
	}
	for _, id := range ids {
		BuffFilter[id+1000000] = true
	}
}

type BuffList struct {
	buffs []*Buff
}

func NewBuffList() *BuffList {
	return &BuffList{
		buffs: make([]*Buff, 0, 24),
	}
}

func (bl *BuffList) Update(now int64) {
	for _, b := range bl.buffs {
		b.UpdateRemain(now)
	}
}

func (bl *BuffList) SetBuffs(buffs []*Buff) {
	filtered := []*Buff{}

	for _, b := range buffs {
		if BuffFilter[b.ID] {
			continue
		}

		filtered = append(filtered, b)
	}

	bl.buffs = append(bl.buffs, filtered...)
}

func (bl *BuffList) Buffs() []*Buff {
	// order buffs by type, debuff first
	sort.Slice(bl.buffs, func(i, j int) bool {
		return bl.buffs[i].Type > bl.buffs[j].Type
	})

	return bl.buffs
}

func (bl *BuffList) DeBuffs() (ret []*Buff) {
	for _, b := range bl.buffs {
		if b.Type == Debuff {
			ret = append(ret, b)
		}
	}

	return
}

func (bl *BuffList) Add(buff *Buff) {
	if BuffFilter[buff.ID] {
		return
	}

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

func (bl *BuffList) UpdateStack(id int64, stack int) {
	for i := range bl.buffs {
		if bl.buffs[i].ID == id {
			bl.buffs[i].Stacks = stack

			return
		}
	}
}

func (bl *BuffList) Refresh(buff *Buff) {
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

func (bl *BuffList) Remove(buff *Buff) {
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
	toRemove := make([]*Buff, 0)

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
	bl.buffs = bl.buffs[:0]
}

type Buff struct {
	Type   BuffType
	ID     int64
	Name   string
	Icon   string
	Stacks int
	Remain int64
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

func (b *Buff) UpdateRemain(now int64) {
	if b.Duration == 0 || b.Duration > 7200*1000 {
		b.Remain = 0

		return
	}

	b.Remain = (b.Duration - util.TickToMS(now-b.ApplyTick)) / 1000
}

func (b Buff) Expired(now int64) bool {
	if b.Duration == 0 {
		return false
	}

	return b.Duration < util.TickToMS(now-b.ApplyTick)
}

func (b Buff) Texture() *ebiten.Image {
	return texture.NewAbilityTexture(b.Icon)
}

func (b *Buff) Remove() {
	if b.RemoveCallback != nil {
		b.RemoveCallback(b, b.ECS, b.Source, b.Target)
	}
}
