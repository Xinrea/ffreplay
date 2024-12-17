package skills

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type SkillTestFan struct {
	DefaultSkill
}

var _ model.GameSkill = (*SkillTestFan)(nil)

func NewSkillTestFan(ecs *ecs.ECS, caster *donburi.Entry, target *donburi.Entry, angle float64) *SkillTestFan {
	casterObj := component.Sprite.Get(caster)
	targetObj := component.Sprite.Get(target)
	casterPos := casterObj.Object.Position()
	targetPos := targetObj.Object.Position()
	direction := targetPos.Sub(casterPos).Radian()
	rangeObj := object.NewFanObject(object.DefaultNegativeSkillRangeOption, casterObj.Object.Position(), angle, 2000)
	rangeObj.Rotate(direction)
	return &SkillTestFan{
		DefaultSkill: DefaultSkill{
			ecs:      ecs,
			caster:   caster,
			target:   target,
			rangeObj: rangeObj,
		},
	}
}

func (s *SkillTestFan) Name() string {
	return "TestRect"
}

func (s *SkillTestFan) InRange(target *donburi.Entry) bool {
	targetObj := component.Sprite.Get(target).Object
	return s.rangeObj.IsPointInside(targetObj.Position())
}

func (s *SkillTestFan) Effect(target *donburi.Entry) {
	status := component.Status.Get(target)
	if status.IsDead() {
		return
	}
	dmg := model.Damage{
		Type:   model.Magical,
		Amount: 10000,
	}
	status.TakeDamage(dmg)
}
