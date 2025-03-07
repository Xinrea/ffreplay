package model

import "github.com/Xinrea/ffreplay/util"

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

	// entries for SkillEvent
	CasterID       int64                  `yaml:"caster"`
	CasterInstance int                    `yaml:"casterins"`
	TargetID       int64                  `yaml:"target"`
	TargetInstance int                    `yaml:"targetins"`
	Offset         int64                  `yaml:"offset"`
	SkillTemplate  string                 `yaml:"skill"`
	SkillConfig    SkillTemplateConfigure `yaml:"config"`

	// entries for MoveEvent
	Duration int64   `yaml:"duration"`
	X        float64 `yaml:"x"`
	y        float64 `yaml:"y"`
	Facing   float64 `yaml:"facing"`
	Lerp     bool    `yaml:"lerp"`

	// entries for MapEvent
	MapUrl  string  `yaml:"map"`
	Scale   float64 `yaml:"scale"`
	OffsetX float64 `yaml:"offsetx"`
	OffsetY float64 `yaml:"offsety"`

	Started bool `yaml:"-"`
}

func (e Event) OffsetTick() int64 {
	return util.MSToTick(e.Offset)
}
