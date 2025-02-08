package skills

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
)

func TemplateFixedRangeSkill(id int64, name string, cast int64, rangeObj object.Object) *model.Skill {
	return &model.Skill{
		ID:          id,
		Name:        name,
		StartTick:   -1,
		Cast:        cast,
		Recast:      0,
		IsGCD:       false,
		EffectRange: rangeObj,
		Initialize:  generalInitializer,
	}
}

func generalInitializer(r object.Object, inst *model.Instance) {
	if r == nil || inst == nil {
		return
	}

	r.UpdateRotate(inst.Face)
	r.UpdatePosition(inst.Object.Position())
}
