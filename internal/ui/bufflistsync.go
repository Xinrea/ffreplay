package ui

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
)

// UIBuff is a display snapshot copied from model.Buff for UI rendering.
type UIBuff struct {
	Type   model.BuffType
	ID     int64
	Name   string
	Icon   string
	Stacks int
	Remain int64
}

func (b *UIBuff) Texture() *ebiten.Image {
	return texture.NewAbilityTexture(b.Icon)
}

// UIBuffList holds the last UI sync snapshot for a model.BuffList.
type UIBuffList struct {
	source *model.BuffList
	buffs  []*UIBuff
}

type buffListUISync struct {
	lastDisplaySecond int64
	lists             map[*model.BuffList]*UIBuffList
}

var buffListUI = &buffListUISync{
	lastDisplaySecond: -1,
	lists:             make(map[*model.BuffList]*UIBuffList),
}

// SyncBuffLists copies every buffable entity's model.BuffList into its UI
// snapshot once per global display second.
func SyncBuffLists(now int64) {
	if ecsInstance == nil {
		return
	}

	displaySecond := util.QuantizeTickToSecond(now)
	if displaySecond == buffListUI.lastDisplaySecond {
		return
	}
	buffListUI.lastDisplaySecond = displaySecond

	seen := make(map[*model.BuffList]struct{})
	for e := range tag.Buffable.Iter(ecsInstance.World) {
		source := component.Status.Get(e).BuffList
		if source == nil {
			continue
		}
		seen[source] = struct{}{}
		buffListUI.listFor(source).sync(displaySecond)
	}

	for source := range buffListUI.lists {
		if _, ok := seen[source]; ok {
			continue
		}
		buffListUI.lists[source].sync(displaySecond)
	}
}

func UIBuffsFor(source *model.BuffList) []*UIBuff {
	if source == nil {
		return nil
	}

	return buffListUI.listFor(source).buffs
}

func UIDebuffsFor(source *model.BuffList) []*UIBuff {
	buffs := UIBuffsFor(source)
	if len(buffs) == 0 {
		return nil
	}

	ret := make([]*UIBuff, 0, len(buffs))
	for _, b := range buffs {
		if b == nil || b.Type != model.Debuff {
			continue
		}
		ret = append(ret, b)
	}

	return ret
}

func (m *buffListUISync) listFor(source *model.BuffList) *UIBuffList {
	list, ok := m.lists[source]
	if !ok {
		list = &UIBuffList{source: source}
		m.lists[source] = list
	}

	return list
}

func (l *UIBuffList) sync(displayNow int64) {
	if l.source == nil {
		l.buffs = nil

		return
	}

	src := l.source.Buffs()
	l.buffs = make([]*UIBuff, 0, len(src))
	for _, b := range src {
		if b == nil {
			continue
		}
		l.buffs = append(l.buffs, cloneUIBuff(b, displayNow))
	}
}

func cloneUIBuff(b *model.Buff, displayNow int64) *UIBuff {
	remain := int64(0)
	if b.Duration > 0 && b.Duration <= 7200*1000 {
		remain = (b.Duration - util.TickToMS(displayNow-b.ApplyTick)) / 1000
	}

	return &UIBuff{
		Type:   b.Type,
		ID:     b.ID,
		Name:   b.Name,
		Icon:   b.Icon,
		Stacks: b.Stacks,
		Remain: remain,
	}
}
