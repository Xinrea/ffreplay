package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/util"
)

func (s *System) SkillUpdate() {
	global := entry.GetGlobal()
	if !global.Loaded.Load() {
		return
	}

	gameObjects := entry.GetGameObjects()
	for _, e := range gameObjects {
		status := component.Status.Get(e)

		for _, instance := range status.Instances {
			casting := instance.GetCast()
			if casting == nil {
				continue
			}

			if casting.StartTick == -1 {
				casting.StartTick = entry.GetTick()
			}

			casting.Update()

			if util.TickToMS(entry.GetTick()-casting.StartTick) > casting.Cast {
				// remove skill
				instance.DoneCast()
			}
		}
	}
}

func (s *System) Cast(
	caster *model.Instance,
	target *model.Instance,
	skill *model.Skill,
	tick int64,
) {
	skill.Init(caster, target)
	caster.Cast(skill)
}
