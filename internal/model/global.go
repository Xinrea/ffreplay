package model

import "sync/atomic"

type GlobalData struct {
	// Real tick * 10 for fine speed control
	Tick          int64
	Speed         int64
	FightDuration atomic.Int64
	Loaded        atomic.Bool
	LoadCount     atomic.Int32
}
