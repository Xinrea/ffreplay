package skills

import (
	"github.com/Xinrea/ffreplay/internal/model"
)

const METER = 25

const (
	SKillCyclonicBreak = 40145
)

// QuerySkill returns a skill from the skillDB, which contains more detailed event timeline to action.
func QuerySkill(skill model.Skill) model.Skill {
	switch skill.ID {
	case SKillCyclonicBreak:
		return NewCyclonicBreak()
	default:
		return skill
	}
}
