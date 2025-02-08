package skills

import (
	"strings"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
)

const METER = 25

// QueryCastingSkill returns a skill from the skillDB, which contains more detailed event timeline to action.
func QueryCastingSkill(skill model.Skill) *model.Skill {
	switch skill.ID {
	case S_Witchgleam:
		return NewWitchGleam(skill.Cast)
	case S_FulminousFieldA:
		return NewFulminousFieldA(skill.Cast)
	case S_FulminousFieldB:
		return NewFulminousFieldB(skill.Cast)
	case S_SidewiseSparkR:
		return NewSideWiseSparkR(skill.Cast)
	case S_SidewiseSparkR2:
		return NewSideWiseSparkR(skill.Cast)
	case S_SidewiseSparkL:
		return NewSideWiseSparkL(skill.Cast)
	case S_SidewiseSparkL2:
		return NewSideWiseSparkL(skill.Cast)
	case S_WitchHunt:
		return NewWitchHunt(skill.Cast)
	case S_LightningCage:
		return NewLightningCage(skill.Cast)
	case S_BewitchingFlight:
		return NewBewitchingFlight(skill.Cast)
	case S_Burst:
		return NewBurst(skill.Cast)
	case S_Thundering:
		return NewThundering(skill.Cast)
	case S_LightningVortex:
		return NewLightningVortex(skill.Cast)
	default:
		actionInfo := model.GetAction(skill.ID)
		if actionInfo == nil {
			return &skill
		}

		if !strings.HasPrefix(actionInfo.Name, "_rsv") && actionInfo.Name != "" {
			skill.Name = actionInfo.Name
		}

		skill.IsGCD = actionInfo.IsGCD

		return &skill
	}
}

func QuerySkill(skill model.Skill) *model.Skill {
	actionInfo := model.GetAction(skill.ID)
	if actionInfo == nil {
		return &skill
	}

	if !strings.HasPrefix(actionInfo.Name, "_rsv") && actionInfo.Name != "" {
		skill.Name = actionInfo.Name
	}

	skill.IsGCD = actionInfo.IsGCD

	return &skill
}

func differByCastTime(cast int64) (object.ObjectOption, int64) {
	opt := object.DefaultNegativeSkillRangeOption
	if cast == 0 {
		opt = object.DefaultEffectSkillRangeOption
		cast = 500
	}

	return opt, cast
}
