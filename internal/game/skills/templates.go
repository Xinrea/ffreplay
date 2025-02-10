package skills

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/vector"
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

func generalInitializer(r object.Object, facing float64, pos vector.Vector) {
	if r == nil {
		return
	}

	r.UpdateRotate(facing)
	r.UpdatePosition(pos)
}
