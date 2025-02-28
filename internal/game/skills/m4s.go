package skills

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
)

const (
	S_LightningCage    = 38351
	S_BewitchingFlight = 38377
	S_Burst            = 38378
	S_SidewiseSparkR   = 38380
	S_SidewiseSparkL   = 38381
	S_SidewiseSparkR2  = 38441
	S_SidewiseSparkL2  = 38442
	S_Thundering       = 19730
	S_LightningVortex  = 19729
	S_WitchHunt        = 38372
	S_Witchgleam       = 38790
	S_FulminousFieldA  = 37118
	S_FulminousFieldB  = 39117
	S_PositronStream   = 38437
	S_NegatronStream   = 38438
)

func NewNegatronStream(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_NegatronStream,
		Name:     "Negatron Stream",
		Cast:     cast,
		Range:    model.RangeTypeRect,
		RangeOpt: opt,
		Width:    12,
		Height:   40,
	}

	return model.TemplateFixedRangeSkill(config)
}

func NewPositronStream(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_PositronStream,
		Name:     "Positron Stream",
		Cast:     cast,
		Range:    model.RangeTypeRect,
		RangeOpt: opt,
		Width:    12,
		Height:   40,
	}

	return model.TemplateFixedRangeSkill(config)
}

func NewSideWiseSparkR(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_SidewiseSparkR,
		Name:     "Sidewise Spark",
		Cast:     cast,
		Range:    model.RangeTypeRect,
		Anchor:   object.AnchorLeftMiddle,
		RangeOpt: opt,
		Width:    20,
		Height:   60,
	}

	skill := model.TemplateFixedRangeSkill(config)

	return skill
}

func NewSideWiseSparkL(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_SidewiseSparkL,
		Name:     "Sidewise Spark",
		Cast:     cast,
		Range:    model.RangeTypeRect,
		Anchor:   object.AnchorRightMiddle,
		RangeOpt: opt,
		Width:    20,
		Height:   60,
	}

	skill := model.TemplateFixedRangeSkill(config)

	return skill
}

func NewFulminousFieldA(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_FulminousFieldA,
		Name:     "Fulminous Field",
		Cast:     cast,
		Range:    model.RangeTypeFan,
		RangeOpt: opt,
		Angle:    22.5,
		Radius:   40,
	}

	return model.TemplateFixedRangeSkill(config)
}

func NewFulminousFieldB(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_FulminousFieldB,
		Name:     "Fulminous Field",
		Cast:     cast,
		Range:    model.RangeTypeFan,
		RangeOpt: opt,
		Angle:    22.5,
		Radius:   40,
	}

	return model.TemplateFixedRangeSkill(config)
}

func NewWitchGleam(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_Witchgleam,
		Name:     "Witch Gleam",
		Cast:     cast,
		Range:    model.RangeTypeRect,
		RangeOpt: opt,
		Width:    4,
		Height:   40,
	}

	return model.TemplateFixedRangeSkill(config)
}

func NewWitchHunt(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_WitchHunt,
		Name:     "Witch Hunt",
		Cast:     cast,
		Range:    model.RangeTypeCircle,
		RangeOpt: opt,
		Radius:   4,
	}

	return model.TemplateFixedRangeSkill(config)
}

func NewLightningCage(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_LightningCage,
		Name:     "Lightning Cage",
		Cast:     cast,
		Range:    model.RangeTypeRect,
		RangeOpt: opt,
		Width:    8,
		Height:   8,
	}

	return model.TemplateFixedRangeSkill(config)
}

func NewBewitchingFlight(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_BewitchingFlight,
		Name:     "Bewitching Flight",
		Cast:     cast,
		Range:    model.RangeTypeRect,
		RangeOpt: opt,
		Width:    4,
		Height:   40,
	}

	return model.TemplateFixedRangeSkill(config)
}

func NewBurst(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_Burst,
		Name:     "Burst",
		Cast:     cast,
		Range:    model.RangeTypeRect,
		RangeOpt: opt,
		Width:    16,
		Height:   40,
	}

	return model.TemplateFixedRangeSkill(config)
}

func NewThundering(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:          S_Thundering,
		Name:        "Thundering",
		Cast:        cast,
		Range:       model.RangeTypeRing,
		RangeOpt:    opt,
		InnerRadius: 10,
		Radius:      30,
	}

	return model.TemplateFixedRangeSkill(config)
}

func NewLightningVortex(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := model.SkillTemplateConfigure{
		ID:       S_LightningVortex,
		Name:     "Lightning Vortex",
		Cast:     cast,
		Range:    model.RangeTypeCircle,
		RangeOpt: opt,
		Radius:   10,
	}

	return model.TemplateFixedRangeSkill(config)
}
