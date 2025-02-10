package skills

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/vector"
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

	skill := TemplateFixedRangeSkill(S_Unknown9cb3, "Unknown9cb3", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		60*METER,
		5*METER,
	))

	return skill
}

func NewThePathOfLight(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	skill := TemplateFixedRangeSkill(S_ThePathOfLight, "The Path of Light", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		60*METER,
		5*METER,
	))

	return skill
}

func NewThePathOfDarkness(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	skill := TemplateFixedRangeSkill(S_ThePathOfDarkness, "The Path of Darkness", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		60*METER,
		5*METER,
	))

	return skill
}

func NewTidalLightA(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	skill := TemplateFixedRangeSkill(S_TidalLightA, "Tidal Light", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		60*METER,
		10*METER,
	))

	return skill
}

func NewTidalLightB(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	skill := TemplateFixedRangeSkill(S_TidalLightB, "Tidal Light", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		60*METER,
		10*METER,
	))

	return skill
}

func NewHallowedWingsL(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	skill := TemplateFixedRangeSkill(S_HallowedWingsL, "allowed Wings", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		60*METER,
		20*METER,
	))

	skill.Initialize = func(r object.Object, facing float64, pos vector.Vector) {
		generalInitializer(r, facing, pos)
		r.UpdateRotate(facing - math.Pi/2)
	}

	return skill
}

func NewHallowedWingsR(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	skill := TemplateFixedRangeSkill(S_HallowedWingsR, "allowed Wings", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		60*METER,
		20*METER,
	))

	skill.Initialize = func(r object.Object, facing float64, pos vector.Vector) {
		generalInitializer(r, facing, pos)
		r.UpdateRotate(facing + math.Pi/2)
	}

	return skill
}

func NewApocalypse(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_Apocalypse, "Apocalypse", cast, object.NewCircleObject(
		opt,
		vector.Vector{},
		9*METER,
	))
}

func NewSinboundMeltdownA(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_SinboundMeltdownA, "Sinbound Meltdown", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		4*METER,
		40*METER,
	))
}

func NewSinboundMeltdownB(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_SinboundMeltdownB, "Sinbound Meltdown", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		4*METER,
		40*METER,
	))
}

func NewICicleImpact(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_ICicleImpact, "ICicle Impact", cast, object.NewCircleObject(
		opt,
		vector.Vector{},
		10*METER,
	))
}

func NewTheHouseOfLight(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_TheHouseOfLight, "The House of Light", cast, object.NewFanObject(
		opt,
		vector.Vector{},
		22.5,
		40*METER,
	))
}

func NewBowShock(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_BowShock, "Bow Shock", cast, object.NewFanObject(
		opt,
		vector.Vector{},
		120,
		40*METER,
	))
}

func NewSinblaze(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(s_Sinblaze, "Sinblaze", cast, object.NewFanObject(
		opt,
		vector.Vector{},
		90,
		40*METER,
	))
}

func NewCyclonicBreakA(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_CyclonicBreakA, "Cyclonic Break", cast, object.NewFanObject(
		opt,
		vector.Vector{},
		22.5,
		40*METER,
	))
}

func NewCyclonicBreakB(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_CyclonicBreakB, "Cyclonic Break", cast, object.NewFanObject(
		opt,
		vector.Vector{},
		22.5,
		40*METER,
	))
}
