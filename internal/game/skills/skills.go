package skills

import (
	"strings"

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
		actionInfo := model.GetAction(skill.ID)
		if actionInfo == nil {
			return skill
		}

		if !strings.HasPrefix(actionInfo.Name, "_rsv") && actionInfo.Name != "" {
			skill.Name = actionInfo.Name
		}

		skill.IsGCD = actionInfo.IsGCD

		return skill
	}
}
