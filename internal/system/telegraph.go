package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// TelegraphUpdate removes expired telegraph entities.
func (s *System) TelegraphUpdate(ecs *ecs.ECS) {
	entriesToRemove := []*donburi.Entry{}

	for e := range tag.Telegraph.Iter(ecs.World) {
		td := component.Telegraph.Get(e)
		if td == nil {
			continue
		}

		// Duration 0 means permanent
		if td.Duration > 0 {
			td.Duration--
			if td.Duration <= 0 {
				entriesToRemove = append(entriesToRemove, e)
			}
		}
	}

	for _, e := range entriesToRemove {
		e.Remove()
	}
}
