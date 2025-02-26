package skills

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/vector"
)

type RangeType int

const (
	RangeTypeRect RangeType = iota
	RangeTypeCircle
	RangeTypeFan
	RangeTypeRing
)

type TemplateConfigure struct {
	ID          int64
	Name        string
	Cast        int64
	Range       RangeType
	RangeOpt    object.ObjectOption
	Anchor      int
	Width       int
	Height      int
	Radius      int
	InnerRadius int
	Angle       float64
}

type SkillTemplateConstructor func(configure TemplateConfigure) *model.Skill

var SkillTemplates = map[string]SkillTemplateConstructor{
	"TemplateFixedRangeSkill": TemplateFixedRangeSkill,
}

func TemplateFixedRangeSkill(configure TemplateConfigure) *model.Skill {
	return &model.Skill{
		ID:          configure.ID,
		Name:        configure.Name,
		StartTick:   -1,
		Cast:        configure.Cast,
		Recast:      0,
		IsGCD:       false,
		EffectRange: rangeObjectFromConfig(configure),
		Initialize:  generalInitializer,
	}
}

func rangeObjectFromConfig(configure TemplateConfigure) object.Object {
	switch configure.Range {
	case RangeTypeRect:
		return object.NewRectObject(
			configure.RangeOpt,
			vector.Vector{},
			configure.Anchor,
			float64(configure.Width*METER),
			float64(configure.Height*METER),
		)
	case RangeTypeCircle:
		return object.NewCircleObject(
			configure.RangeOpt,
			vector.Vector{},
			float64(configure.Radius*METER),
		)
	case RangeTypeFan:
		return object.NewFanObject(
			configure.RangeOpt,
			vector.Vector{},
			float64(configure.Radius*METER),
			configure.Angle,
		)
	case RangeTypeRing:
		return object.NewRingObject(
			configure.RangeOpt,
			vector.Vector{},
			float64(configure.InnerRadius*METER),
			float64(configure.Radius*METER),
		)
	default:
		return nil
	}
}

func generalInitializer(r object.Object, facing float64, pos vector.Vector) {
	if r == nil {
		return
	}

	r.UpdateRotate(facing)
	r.UpdatePosition(pos)
}
