package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
)

func (s *System) BuffUpdate(tick int64) {
	global := entry.GetGlobal()
	if !global.Loaded.Load() {
		return
	}

	buffables := entry.GetBuffables()

	for _, e := range buffables {
		component.Status.Get(e).BuffList.Update(tick)
	}

	// in replay mode, buff expires are ctrolled by logs events.
	if entry.GetGlobal().ReplayMode {
		return
	}

	for _, e := range buffables {
		component.Status.Get(e).BuffList.Update(tick)
	}
}
