package skills

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/vector"
)

const (
	S_CyclonicBreakA = 40145
	S_CyclonicBreakB = 40146
)

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
