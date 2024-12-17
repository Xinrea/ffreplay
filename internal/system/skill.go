package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/donburi/ecs"
)

func (s *System) SkillUpdate(ecs *ecs.ECS) {
	for e := range tag.Skill.Iter(ecs.World) {
		skill := component.Skill.Get(e)
		skill.GameSkill.Update()
		// skill target is setted at cast time
		if util.TickToMS(entry.GetTick(ecs)-skill.Time.StartTick) >= skill.Time.CastTime {
			// if in replay mode, no effect will be applied
			if !s.InReplay {
				// find all target in range
				for e := range tag.PartyMember.Iter(ecs.World) {
					status := component.Status.Get(e)
					if status.IsDead() {
						continue
					}
					if skill.GameSkill.InRange(e) {
						skill.GameSkill.Effect(e)
					}
				}
			}
			// remove skill
			ecs.World.Remove(e.Entity())
		}
	}
}
