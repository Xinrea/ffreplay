package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func (s *System) SkillUpdate(ecs *ecs.ECS) {
	for e := range tag.GameObject.Iter(ecs.World) {
		sprite := component.Sprite.Get(e)
		if !sprite.Initialized {
			continue
		}
		for _, instance := range sprite.Instances {
			if instance.Casting == nil {
				continue
			}
			if instance.Casting.StartTick == -1 {
				instance.Casting.StartTick = entry.GetTick(ecs)
			}
			if util.TickToMS(entry.GetTick(ecs)-instance.Casting.StartTick) >= instance.Casting.Cast {
				// remove skill
				instance.Casting = nil
			}
		}
	}
}

func (s *System) Cast(ecs *ecs.ECS, caster *donburi.Entry, casterInstance int, target *donburi.Entry, targetInstance int, skill model.Skill) {
	global := component.Global.Get(tag.Global.MustFirst(ecs.World))
	casterSprite := component.Sprite.Get(caster)
	if casterSprite == nil {
		return
	}
	if casterInstance >= len(casterSprite.Instances) {
		return
	}
	casterSprite.Instances[casterInstance].Cast(skill)
	if skill.SkillEvents != nil && global.Debug {
		timeline := skill.SkillEvents.InstanceWith(entry.GetTick(ecs), caster, casterInstance, target, targetInstance)
		entry.NewTimeline(ecs, &timeline)
	}
}
