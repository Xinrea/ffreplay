package model

import (
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type TimelineData struct {
	Name           string
	StartTick      int64
	Caster         *donburi.Entry
	CasterInstance int
	Target         *donburi.Entry
	TargetInstance int
	Events         []Event
}

func (t TimelineData) InstanceWith(tick int64, caster *donburi.Entry, casterInstance int, target *donburi.Entry, targetInstance int) TimelineData {
	t.StartTick = tick
	t.Caster = caster
	t.CasterInstance = casterInstance
	t.Target = target
	t.TargetInstance = targetInstance
	return t
}

func (t TimelineData) IsDone(tick int64) bool {
	if len(t.Events) == 0 {
		return true
	}
	d := tick - t.StartTick
	return d > t.EndTick() || d < 0
}

func (t TimelineData) EndTick() int64 {
	return util.MSToTick(t.Events[len(t.Events)-1].Offset + t.Events[len(t.Events)-1].DisplayTime)
}

func (t TimelineData) Begin(ecs *ecs.ECS, index int) {
	if t.Events[index].Begin == nil {
		return
	}
	t.Events[index].Begin(ecs, t.Events[index].EffectRange, t.Caster, t.CasterInstance, t.Target, t.TargetInstance)
}

func (t TimelineData) Finish(ecs *ecs.ECS, index int) {
	if t.Events[index].Finish == nil {
		return
	}
	t.Events[index].Finish(ecs, t.Events[index].EffectRange, t.Caster, t.CasterInstance, t.Target, t.TargetInstance)
}

func (t TimelineData) Update(ecs *ecs.ECS, index int) {
	if t.Events[index].Update == nil {
		return
	}
	t.Events[index].Update(ecs, t.Events[index].EffectRange, t.Caster, t.CasterInstance, t.Target, t.TargetInstance)
}

func (t *TimelineData) Reset() {
	t.StartTick = 0
}

type EventCallback func(ecs *ecs.ECS, rangeObj object.Object, caster *donburi.Entry, casterInstance int, target *donburi.Entry, targetInstance int)

type Event struct {
	Offset      int64
	DisplayTime int64
	EffectRange object.Object
	Begin       EventCallback
	Update      EventCallback
	Finish      EventCallback
}

func (e Event) OffsetTick() int64 {
	return util.MSToTick(e.Offset)
}

func (e Event) DisplayTick() int64 {
	return util.MSToTick(e.DisplayTime)
}
