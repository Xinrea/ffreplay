package skills

import (
	"strings"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
)

const METER = 25

// QueryCastingSkill returns a skill from the skillDB, which contains more detailed event timeline to action.
var skillHandlers = map[int64]func(cast int64) *model.Skill{
	S_Witchgleam:       NewWitchGleam,
	S_FulminousFieldA:  NewFulminousFieldA,
	S_FulminousFieldB:  NewFulminousFieldB,
	S_SidewiseSparkR:   NewSideWiseSparkR,
	S_SidewiseSparkR2:  NewSideWiseSparkR,
	S_SidewiseSparkL:   NewSideWiseSparkL,
	S_SidewiseSparkL2:  NewSideWiseSparkL,
	S_WitchHunt:        NewWitchHunt,
	S_LightningCage:    NewLightningCage,
	S_BewitchingFlight: NewBewitchingFlight,
	S_Burst:            NewBurst,
	S_Thundering:       NewThundering,
	S_LightningVortex:  NewLightningVortex,
}

func QueryCastingSkill(skill model.Skill) *model.Skill {
	if handler, found := skillHandlers[skill.ID]; found {
		return handler(skill.Cast)
	}

	return handleDefaultSkill(skill)
}

func handleDefaultSkill(skill model.Skill) *model.Skill {
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
