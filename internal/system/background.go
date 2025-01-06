package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
)

func (s *System) BackgroundUpdate() {
	global := entry.GetGlobal(s.ecs)
	if !global.Loaded.Load() {
		return
	}

	ground := component.Map.Get(component.Map.MustFirst(s.ecs.World))
	// only auto update phase in replay mode
	if global.ReplayMode {
		if len(ground.Config.Phases) > 0 {
			// find current phase
			p := entry.GetPhase(s.ecs)
			if p < 0 || p >= len(ground.Config.Phases) {
				p = 0
			}

			ground.Config.CurrentPhase = p
		}
	}
}
