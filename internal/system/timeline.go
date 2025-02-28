package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/yohamta/donburi/ecs"
)

func (s *System) TimelineUpdate(ecs *ecs.ECS) {
	global := entry.GetGlobal(ecs)
	if !global.Loaded.Load() {
		return
	}

	tick := entry.GetTick(ecs)

	for e := range tag.Timeline.Iter(ecs.World) {
		timeline := component.Timeline.Get(e)
		if timeline.IsDone(tick) {
			e.Remove()

			continue
		}

		p := tick - timeline.StartTick
		for i := range timeline.Events {
			updateEvent(timeline, i, p, ecs)
		}
	}
}

func updateEvent(timeline *model.TimelineData, i int, p int64, ecs *ecs.ECS) {
	event := timeline.Events[i]
	if !event.Started && p >= event.OffsetTick() {
		timeline.Begin(ecs, i)

		// handle this event
		skill := model.SkillTemplates[event.SkillTemplate](event.SkillConfig)
		if skill == nil {
			return
		}

		caster := entry.GetStatusByID(ecs, event.CasterID)
		target := entry.GetStatusByID(ecs, event.TargetID)

		skill.Init(caster.Instances[event.CasterInstance], target.Instances[event.TargetInstance])

		caster.Instances[0].Cast(skill)

		return
	}
}
