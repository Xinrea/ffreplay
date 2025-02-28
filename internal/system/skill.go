package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/donburi/ecs"
)

func (s *System) SkillUpdate(ecs *ecs.ECS) {
	global := entry.GetGlobal(ecs)
	if !global.Loaded.Load() {
		return
	}

	for e := range tag.GameObject.Iter(ecs.World) {
		status := component.Status.Get(e)

		for _, instance := range status.Instances {
			casting := instance.GetCast()
			if casting == nil {
				continue
			}

			if casting.StartTick == -1 {
				casting.StartTick = entry.GetTick(ecs)
			}

			casting.Update()

			if util.TickToMS(entry.GetTick(ecs)-casting.StartTick) > casting.Cast {
				// remove skill
				instance.DoneCast()
			}
		}
	}
}

func (s *System) Cast(
	ecs *ecs.ECS,
	caster *model.Instance,
	target *model.Instance,
	skill *model.Skill,
	tick int64,
) {
	skill.Init(caster, target)
	caster.Cast(skill)
}
