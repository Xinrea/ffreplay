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
	if !timeline.Events[i].Started && p >= timeline.Events[i].OffsetTick() {
		timeline.Begin(ecs, i)

		return
	}

	if timeline.Events[i].OffsetTick() < p && p < timeline.Events[i].OffsetTick()+timeline.Events[i].DurationTick() {
		timeline.Update(ecs, i)

		return
	}

	if !timeline.Events[i].Finished && p >= timeline.Events[i].OffsetTick()+timeline.Events[i].DurationTick() {
		timeline.Finish(ecs, i)

		return
	}
}
