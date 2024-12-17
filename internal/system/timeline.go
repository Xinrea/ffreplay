package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/yohamta/donburi/ecs"
)

func (s *System) TimelineUpdate(ecs *ecs.ECS) {
	for e := range tag.Timeline.Iter(ecs.World) {
		timeline := component.Timeline.Get(e)
		for _, event := range timeline.Events {
			if event.Done {
				continue
			}
			if event.Offset <= entry.GetTick(ecs) {
				event.Action(ecs)
				event.Done = true
			}
		}
	}
}
