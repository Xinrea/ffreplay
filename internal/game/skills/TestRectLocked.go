package skills

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type SkillTestRectLocked struct {
	DefaultSkill
}

var _ model.GameSkill = (*SkillTestRectLocked)(nil)

func NewSkillTestRectLocked(ecs *ecs.ECS, caster *donburi.Entry, target *donburi.Entry, width int) *SkillTestRectLocked {
	return &SkillTestRectLocked{
		DefaultSkill: DefaultSkill{
			ecs:      ecs,
			caster:   caster,
			target:   target,
			rangeObj: object.NewRectObject(object.DefaultNegativeSkillRangeOption, vector.Vector{}, float64(width), 5000),
		},
	}
}

func (s *SkillTestRectLocked) Name() string {
	return "TestRectLocked"
}

func (s *SkillTestRectLocked) Update() {
	// update radian using caster and target position
	caster := component.Sprite.Get(s.caster)
	target := component.Sprite.Get(s.target)
	targetPos := target.Object.Position()
	s.rangeObj.UpdateRotate(targetPos.Sub(caster.Object.Position()).Radian())
}

func (s *SkillTestRectLocked) InRange(target *donburi.Entry) bool {
	targetObj := component.Sprite.Get(target).Object
	return s.rangeObj.IsPointInside(targetObj.Position())
}

func (s *SkillTestRectLocked) Effect(target *donburi.Entry) {
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
