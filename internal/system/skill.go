package system

import (
	"log"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/donburi"
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

			if util.TickToMS(entry.GetTick(ecs)-casting.StartTick) > casting.Cast {
				// remove skill
				instance.DoneCast()
			}

			casting.Update(instance.Face, instance.Object.Position())
		}
	}
}

func (s *System) Cast(
	ecs *ecs.ECS,
	caster *donburi.Entry,
	casterInstance int,
	target *donburi.Entry,
	targetInstance int,
	skill *model.Skill,
	tick int64,
) {
	casterSprite := component.Status.Get(caster)
	if casterSprite == nil {
		log.Println("Cast with nil caster")

		return
	}

	if casterInstance >= len(casterSprite.Instances) {
		log.Println("Cast with invalid caster instance id")

		return
	}

	inst := casterSprite.Instances[casterInstance]

	skill.Initialize(inst.Face, inst.Object.Position())

	inst.Cast(skill)
}
