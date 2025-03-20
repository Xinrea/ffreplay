package system

import (
	"github.com/Xinrea/ffreplay/internal/entry"
)

func (s *System) BackgroundUpdate() {
	global := entry.GetGlobal()
	if !global.Loaded.Load() {
		return
	}

	ground := entry.GetMap()
	// only auto update phase in replay mode
	if global.ReplayMode {
		if len(ground.Config.Phases) > 0 {
			// find current phase
			p := entry.GetPhase()
			if p < 0 || p >= len(ground.Config.Phases) {
				p = 0
			}

			ground.Config.CurrentPhase = p
		}
	}
}
