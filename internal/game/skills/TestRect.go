package skills

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type SkillTestRect struct {
	DefaultSkill
}

var _ model.GameSkill = (*SkillTestRect)(nil)

func NewSkillTestRect(ecs *ecs.ECS, caster *donburi.Entry, target *donburi.Entry, width int) *SkillTestRect {
	casterObj := component.Sprite.Get(caster)
	targetObj := component.Sprite.Get(target)
	targetPos := targetObj.Object.Position()
	rangeObj := object.NewRectObject(object.DefaultNegativeSkillRangeOption, vector.Vector{}, float64(width), 5000)
	radian := targetPos.Sub(casterObj.Object.Position()).Radian()
	rangeObj.Rotate(radian)
	return &SkillTestRect{
		DefaultSkill: DefaultSkill{
			ecs:      ecs,
			caster:   caster,
			target:   target,
			rangeObj: rangeObj,
		},
	}
}

func (s *SkillTestRect) Name() string {
	return "TestRect"
}

func (s *SkillTestRect) InRange(target *donburi.Entry) bool {
	targetObj := component.Sprite.Get(target).Object
	return s.rangeObj.IsPointInside(targetObj.Position())
}

func (s *SkillTestRect) Effect(target *donburi.Entry) {
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
