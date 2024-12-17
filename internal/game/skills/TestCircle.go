package skills

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type SkillTestCircle struct {
	DefaultSkill
}

var _ model.GameSkill = (*SkillTestCircle)(nil)

func NewSkillTestCircle(ecs *ecs.ECS, caster *donburi.Entry, target *donburi.Entry, radius float64) *SkillTestCircle {
	pos := component.Sprite.Get(target).Object.Position()
	rangeObj := object.NewCircleObject(object.DefaultNegativeSkillRangeOption, pos, radius*METER)
	return &SkillTestCircle{
		DefaultSkill: DefaultSkill{
			ecs:      ecs,
			caster:   caster,
			target:   target,
			rangeObj: rangeObj,
		},
	}
}

func (s *SkillTestCircle) Name() string {
	return "TestCircle"
}

func (s *SkillTestCircle) InRange(target *donburi.Entry) bool {
	return s.rangeObj.IsPointInside(component.Sprite.Get(target).Object.Position())
}

func (s *SkillTestCircle) Effect(target *donburi.Entry) {
	status := component.Status.Get(target)
	if status.IsDead() {
		return
	}
	dmg := model.Damage{
		Type:   model.Magical,
		Amount: 50000,
	}
	status.TakeDamage(dmg)
}
