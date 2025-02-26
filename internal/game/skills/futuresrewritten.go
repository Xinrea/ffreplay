package skills

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
)

const (
	S_CyclonicBreakA  = 40145
	S_CyclonicBreakB  = 40146
	S_BowShock        = 40143
	s_Sinblaze        = 40156
	S_ICicleImpact    = 40198
	S_TheHouseOfLight = 40206
	// 40235 Sinbound Meltdown.
	S_SinboundMeltdownA = 40235
	S_SinboundMeltdownB = 40292
	// 40297 Apocalypse.
	S_Apocalypse     = 40297
	S_HallowedWingsL = 40227
	S_HallowedWingsR = 40228
	// 40252 Tidal Light.
	S_TidalLightA = 40252
	S_TidalLightB = 40253
	// 40118 the Path of Darkness.
	S_Unknown9cb3       = 40115
	S_ThePathOfDarkness = 40118
	// 40307 the Path of Light.
	S_ThePathOfLight     = 40307
	S_ThePathOfLightB    = 40308
	S_ThePathOfDarknessB = 40309
)

func NewUnknown9cb3(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_Unknown9cb3,
		Name:     "Unknown9cb3",
		Cast:     cast,
		Range:    RangeTypeRect,
		RangeOpt: opt,
		Width:    60,
		Height:   5,
	}

	skill := TemplateFixedRangeSkill(config)

	return skill
}

func NewThePathOfLight(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_ThePathOfLight,
		Name:     "The Path of Light",
		Cast:     cast,
		Range:    RangeTypeRect,
		RangeOpt: opt,
		Width:    60,
		Height:   5,
	}

	skill := TemplateFixedRangeSkill(config)

	return skill
}

func NewThePathOfDarkness(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_ThePathOfDarkness,
		Name:     "The Path of Darkness",
		Cast:     cast,
		Range:    RangeTypeRect,
		RangeOpt: opt,
		Width:    60,
		Height:   5,
	}

	skill := TemplateFixedRangeSkill(config)

	return skill
}

func NewTidalLightA(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_TidalLightA,
		Name:     "Tidal Light",
		Cast:     cast,
		Range:    RangeTypeRect,
		RangeOpt: opt,
		Width:    60,
		Height:   10,
	}

	skill := TemplateFixedRangeSkill(config)

	return skill
}

func NewTidalLightB(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_TidalLightB,
		Name:     "Tidal Light",
		Cast:     cast,
		Range:    RangeTypeRect,
		RangeOpt: opt,
		Width:    60,
		Height:   10,
	}

	skill := TemplateFixedRangeSkill(config)

	return skill
}

func NewHallowedWingsL(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_HallowedWingsL,
		Name:     "Hallowed Wings",
		Cast:     cast,
		Range:    RangeTypeRect,
		Anchor:   object.AnchorRightMiddle,
		RangeOpt: opt,
		Width:    20,
		Height:   60,
	}

	skill := TemplateFixedRangeSkill(config)

	return skill
}

func NewHallowedWingsR(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_HallowedWingsR,
		Name:     "Hallowed Wings",
		Cast:     cast,
		Range:    RangeTypeRect,
		Anchor:   object.AnchorLeftMiddle,
		RangeOpt: opt,
		Width:    20,
		Height:   60,
	}

	skill := TemplateFixedRangeSkill(config)

	return skill
}

func NewApocalypse(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_Apocalypse,
		Name:     "Apocalypse",
		Cast:     cast,
		Range:    RangeTypeCircle,
		RangeOpt: opt,
		Radius:   9,
	}

	return TemplateFixedRangeSkill(config)
}

func NewSinboundMeltdownA(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_SinboundMeltdownA,
		Name:     "Sinbound Meltdown",
		Cast:     cast,
		Range:    RangeTypeRect,
		RangeOpt: opt,
		Width:    4,
		Height:   40,
	}

	return TemplateFixedRangeSkill(config)
}

func NewSinboundMeltdownB(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_SinboundMeltdownB,
		Name:     "Sinbound Meltdown",
		Cast:     cast,
		Range:    RangeTypeRect,
		RangeOpt: opt,
		Width:    4,
		Height:   40,
	}

	return TemplateFixedRangeSkill(config)
}

func NewICicleImpact(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_ICicleImpact,
		Name:     "ICicle Impact",
		Cast:     cast,
		Range:    RangeTypeCircle,
		RangeOpt: opt,
		Radius:   10,
	}

	return TemplateFixedRangeSkill(config)
}

func NewTheHouseOfLight(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_TheHouseOfLight,
		Name:     "The House of Light",
		Cast:     cast,
		Range:    RangeTypeFan,
		RangeOpt: opt,
		Angle:    22.5,
		Radius:   40,
	}

	return TemplateFixedRangeSkill(config)
}

func NewBowShock(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_BowShock,
		Name:     "Bow Shock",
		Cast:     cast,
		Range:    RangeTypeFan,
		RangeOpt: opt,
		Angle:    120,
		Radius:   40,
	}

	return TemplateFixedRangeSkill(config)
}

func NewSinblaze(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       s_Sinblaze,
		Name:     "Sinblaze",
		Cast:     cast,
		Range:    RangeTypeFan,
		RangeOpt: opt,
		Angle:    90,
		Radius:   40,
	}

	return TemplateFixedRangeSkill(config)
}

func NewCyclonicBreakA(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_CyclonicBreakA,
		Name:     "Cyclonic Break",
		Cast:     cast,
		Range:    RangeTypeFan,
		RangeOpt: opt,
		Angle:    22.5,
		Radius:   40,
	}

	return TemplateFixedRangeSkill(config)
}

func NewCyclonicBreakB(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	config := TemplateConfigure{
		ID:       S_CyclonicBreakB,
		Name:     "Cyclonic Break",
		Cast:     cast,
		Range:    RangeTypeFan,
		RangeOpt: opt,
		Angle:    22.5,
		Radius:   40,
	}

	return TemplateFixedRangeSkill(config)
}
