package model

import (
	"sync/atomic"

	"github.com/yohamta/donburi"
)

type GlobalData struct {
	// Real tick * 10 for fine speed control
	RenderNPC    bool
	ReplayMode   bool
	Tick         int64
	Speed        int64
	TargetPlayer *donburi.Entry
	LimitBreak   int
	Bar          int
	Debug        bool
	RangeDisplay bool
	// should not process any event when UI is on focus
	UIFocus bool
	// UIHovered is true when the cursor is over an interactive UI panel
	// (e.g. the property panel). World interaction is suppressed while true.
	UIHovered bool
	// Selected is the currently selected game object entry in playground mode.
	Selected *donburi.Entry
	// SelectedInstance is the instance index within the selected entry's sprite.
	SelectedInstance int
	// Phases is a tick array for phase change
	Phases         []int64
	FightDuration  atomic.Int64
	Loaded         atomic.Bool
	LoadCount      atomic.Int32
	LoadTotal      int
	ShowTargetRing bool
	// WorldMarker selected
	WorldMarkerSelected int
	// marker for system reset
	Reset atomic.Bool
}
