package skills

import (
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

const METER = 25

type DefaultSkill struct {
	ecs      *ecs.ECS
	caster   *donburi.Entry
	target   *donburi.Entry
	rangeObj object.Object
}

var _ model.GameSkill = (*DefaultSkill)(nil)

func (s *DefaultSkill) Name() string {
	return ""
}

func (s *DefaultSkill) Caster() *donburi.Entry {
	return s.caster
}

func (s *DefaultSkill) Target() *donburi.Entry {
	return s.target
}

func (s *DefaultSkill) Update() {}

func (s *DefaultSkill) Range() object.Object {
	return s.rangeObj
}

func (s *DefaultSkill) InRange(target *donburi.Entry) bool {
	return false
}

func (s *DefaultSkill) Effect(target *donburi.Entry) {}
