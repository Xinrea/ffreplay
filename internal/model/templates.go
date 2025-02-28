package model

import (
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/vector"
)

const METER = 25

type SkillTemplateConstructor func(configure SkillTemplateConfigure) *Skill

var SkillTemplates = map[string]SkillTemplateConstructor{
	"TemplateFixedRangeSkill":  TemplateFixedRangeSkill,
	"TemplateLockedRangeSkill": TemplateLockedRangeSkill,
}

func TemplateFixedRangeSkill(configure SkillTemplateConfigure) *Skill {
	return &Skill{
		ID:          configure.ID,
		Name:        configure.Name,
		StartTick:   -1,
		Cast:        configure.Cast,
		Recast:      0,
		IsGCD:       false,
		EffectRange: rangeObjectFromConfig(configure),
		Initializer: generalInitializer,
		Updater:     nil,
	}
}

func TemplateLockedRangeSkill(configure SkillTemplateConfigure) *Skill {
	return &Skill{
		ID:          configure.ID,
		Name:        configure.Name,
		StartTick:   -1,
		Cast:        configure.Cast,
		Recast:      0,
		IsGCD:       false,
		EffectRange: rangeObjectFromConfig(configure),
		Initializer: lockedSkillHandler,
		Updater:     lockedSkillHandler,
	}
}

func rangeObjectFromConfig(configure SkillTemplateConfigure) object.Object {
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

// generalInitializer is a general initializer for skill effect range.
// It will initialize the effect range's position and facing based on caster only.
func generalInitializer(caster *Instance, target *Instance, effectRange object.Object) {
	if effectRange == nil {
		return
	}

	effectRange.UpdateRotate(caster.Face)
	effectRange.UpdatePosition(caster.Object.Position())
}

// lockedSkillHandler is a general handler for skill effect range.
// It will update the effect range's position and facing based on caster and target.
func lockedSkillHandler(caster *Instance, target *Instance, effectRange object.Object) {
	if effectRange == nil {
		return
	}

	switch effectRange.Type() {
	case object.TypeRect:
		fallthrough
	case object.TypeFan:
		pTarget := target.Object.Position()
		pSelf := caster.Object.Position()
		// facing to the target
		facing := pTarget.Sub(pSelf).Radian()
		effectRange.UpdateRotate(facing + caster.Face)
		effectRange.UpdatePosition(pSelf)
	case object.TypeCircle:
		fallthrough
	case object.TypeRing:
		effectRange.UpdatePosition(target.Object.Position())
	}
}
