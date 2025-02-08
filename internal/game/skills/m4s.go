package skills

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/vector"
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
)

func NewSideWiseSparkR(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	skill := TemplateFixedRangeSkill(S_SidewiseSparkR, "Sidewise Spark", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		60*METER,
		20*METER,
	))

	skill.Initialize = func(r object.Object, inst *model.Instance) {
		generalInitializer(r, inst)
		r.UpdateRotate(inst.Face + math.Pi/2)
	}

	return skill
}

func NewSideWiseSparkL(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	skill := TemplateFixedRangeSkill(S_SidewiseSparkL, "Sidewise Spark", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		60*METER,
		20*METER,
	))

	skill.Initialize = func(r object.Object, inst *model.Instance) {
		generalInitializer(r, inst)
		r.UpdateRotate(inst.Face - math.Pi/2)
	}

	return skill
}

func NewFulminousFieldA(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_FulminousFieldA, "Fulminous Field", cast, object.NewFanObject(
		opt,
		vector.Vector{},
		22.5,
		40*METER,
	))
}

func NewFulminousFieldB(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_FulminousFieldB, "Fulminous Field", cast, object.NewFanObject(
		opt,
		vector.Vector{},
		22.5,
		40*METER,
	))
}

func NewWitchGleam(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_Witchgleam, "Witch Gleam", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		4*METER,
		40*METER,
	))
}

func NewWitchHunt(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_WitchHunt, "Witch Hunt", cast, object.NewCircleObject(
		opt,
		vector.Vector{},
		4*METER,
	))
}

func NewLightningCage(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_LightningCage, "Lightning Cage", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorMiddle,
		8*METER,
		8*METER,
	))
}

func NewBewitchingFlight(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_BewitchingFlight, "Bewitching Flight", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		4*METER,
		40*METER,
	))
}

func NewBurst(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_Burst, "Burst", cast, object.NewRectObject(
		opt,
		vector.Vector{},
		object.AnchorBottomMiddle,
		16*METER,
		40*METER,
	))
}

func NewThundering(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_Thundering, "Thundering", cast, object.NewRingObject(
		opt,
		vector.Vector{},
		10*METER,
		30*METER,
	))
}

func NewLightningVortex(cast int64) *model.Skill {
	opt, cast := differByCastTime(cast)

	return TemplateFixedRangeSkill(S_LightningVortex, "Lightning Vortex", cast, object.NewCircleObject(
		opt,
		vector.Vector{},
		10*METER,
	))
}
