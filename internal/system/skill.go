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
		sprite := component.Sprite.Get(e)
		if !sprite.Initialized {
			continue
		}

		for _, instance := range sprite.Instances {
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
	casterSprite := component.Sprite.Get(caster)
	if casterSprite == nil {
		log.Println("Cast with nil caster")

		return
	}

	if casterInstance >= len(casterSprite.Instances) {
		log.Println("Cast with invalid caster instance id")

		return
	}

	casterSprite.Instances[casterInstance].Cast(skill)
}
