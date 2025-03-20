package model

import (
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/Xinrea/ffreplay/util"
)

type EventType int

const (
	EventTypeSkill EventType = iota
	EventTypeMove
	EventTypeMap
)

// Event is a single skill trigger in the timeline
//
// when world tick >= offset + timeline.StartTick, the event will be triggered
// and the skill will be casted. After this, skill will be handled by skill system.
type Event struct {
	Type EventType `yaml:"type"`
	// OffsetMS is the time when event begins relative to the beginning of timeline
	OffsetMS int64 `yaml:"offset"`

	SkillEvent
	MoveEvent
	MapEvent

	Started bool `yaml:"-"`
}

type SkillEvent struct {
	// entries for SkillEvent
	CasterID       int64                  `yaml:"caster"`
	CasterInstance int                    `yaml:"casterins"`
	TargetID       int64                  `yaml:"target"`
	TargetInstance int                    `yaml:"targetins"`
	SkillTemplate  string                 `yaml:"skill"`
	SkillConfig    SkillTemplateConfigure `yaml:"config"`
}

type MoveEvent struct {
	// entries for MoveEvent
	InitialPosition vector.Vector `yaml:"-"`
	InitialFace     float64       `yaml:"-"`
	TargetID        int64         `yaml:"target"`
	InstanceID      int           `yaml:"instance"`
	DurationMS      int64         `yaml:"duration"`
	X               float64       `yaml:"x"`
	Y               float64       `yaml:"y"`
	Facing          float64       `yaml:"facing"`
}

type MapEvent struct {
	// entries for MapEvent
	MapUrl  string  `yaml:"map"`
	Scale   float64 `yaml:"scale"`
	OffsetX float64 `yaml:"offsetx"`
	OffsetY float64 `yaml:"offsety"`
}

// OffsetTick returns time in tick.
func (e Event) OffsetTick() int64 {
	return util.MSToTick(e.OffsetMS)
}

func (e Event) ProgressedTick(timelineTick, currentTick int64) int64 {
	return currentTick - (timelineTick + e.OffsetTick())
}
