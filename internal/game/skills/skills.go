package skills

import (
	"strings"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
)

// QueryCastingSkill returns a skill from the skillDB, which contains more detailed event timeline to action.
// Some skills are not enabled because they depends on npc's precise face and position
// which is not available in fflogs.
var skillHandlers = map[int64]func(cast int64) *model.Skill{
	S_Witchgleam: NewWitchGleam,
	// S_FulminousFieldA:  NewFulminousFieldA,
	// S_FulminousFieldB:  NewFulminousFieldB,
	S_SidewiseSparkR:  NewSideWiseSparkR,
	S_SidewiseSparkR2: NewSideWiseSparkR,
	S_SidewiseSparkL:  NewSideWiseSparkL,
	S_SidewiseSparkL2: NewSideWiseSparkL,
	// S_WitchHunt:        NewWitchHunt,
	S_LightningCage:    NewLightningCage,
	S_BewitchingFlight: NewBewitchingFlight,
	S_Burst:            NewBurst,
	S_Thundering:       NewThundering,
	S_LightningVortex:  NewLightningVortex,
	// S_CyclonicBreakA:   NewCyclonicBreakA,
	// S_CyclonicBreakB:   NewCyclonicBreakB,
	S_PositronStream: NewPositronStream,
	S_NegatronStream: NewNegatronStream,
	// S_BowShock:         NewBowShock,
	// s_Sinblaze:         NewSinblaze,
	S_ICicleImpact: NewICicleImpact,
	// S_TheHouseOfLight:  NewTheHouseOfLight,
	// S_SinboundMeltdownA: NewSinboundMeltdownA,
	// S_SinboundMeltdownB: NewSinboundMeltdownB,
	S_Apocalypse:     NewApocalypse,
	S_HallowedWingsL: NewHallowedWingsL,
	S_HallowedWingsR: NewHallowedWingsR,
	S_TidalLightA:    NewTidalLightA,
	S_TidalLightB:    NewTidalLightB,
	// S_Unknown9cb3:       NewUnknown9cb3,
	S_ThePathOfLight:    NewThePathOfLight,
	S_ThePathOfDarkness: NewThePathOfDarkness,
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
