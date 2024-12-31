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
	// Phases is a tick array for phase change
	Phases        []int64
	FightDuration atomic.Int64
	Loaded        atomic.Bool
	LoadCount     atomic.Int32
	LoadTotal     int
	// marker for system reset
	Reset atomic.Bool
}
