package model

import (
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type TimelineData struct {
	Name      string `yaml:"name"`
	StartTick int64  `yaml:"-"`
	// Caster is the entity that handles all skills in the timeline
	Caster *donburi.Entry `yaml:"-"`
	Events []Event        `yaml:"events"`
}

func (t *TimelineData) InstanceWith(
	caster *donburi.Entry,
	tick int64,
	targetInstance int,
) *TimelineData {
	t.StartTick = tick
	t.Caster = caster

	return t
}

func (t TimelineData) IsDone(tick int64) bool {
	if len(t.Events) == 0 {
		return true
	}

	isDone := true

	for _, e := range t.Events {
		if !e.Started {
			isDone = false

			break
		}
	}

	return isDone
}

func (t TimelineData) Begin(ecs *ecs.ECS, index int) {
	t.Events[index].Started = true
}

func (t *TimelineData) Reset() {
	t.StartTick = 0
}

type RangeType int

const (
	RangeTypeRect RangeType = iota
	RangeTypeCircle
	RangeTypeFan
	RangeTypeRing
)

type SkillTemplateConfigure struct {
	ID          int64
	Name        string
	Cast        int64
	Range       RangeType
	RangeOpt    object.ObjectOption
	Anchor      int
	Width       int
	Height      int
	Radius      int
	InnerRadius int
	Angle       float64
}
