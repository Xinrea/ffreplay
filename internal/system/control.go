package system

import (
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi/ecs"
)

const MaxVelocity = 4

func (s *System) ControlUpdate(ecs *ecs.ECS) {
	camera := entry.GetCamera(s.ecs)
	camera.Update(s.ViewPort)

	global := entry.GetGlobal(s.ecs)
	if !global.Loaded.Load() {
		return
	}

	if global.UIFocus {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackquote) {
		global.Debug = !global.Debug
	}

	_, dy := ebiten.Wheel()

	if util.IsWasm() {
		camera.ZoomFactor -= int(dy)
	} else {
		camera.ZoomFactor -= int(dy * 3)
	}

	if global.ReplayMode || global.TargetPlayer == nil {
		s.replayModeControl(ecs)
	} else {
		s.playgroundControl(ecs)
	}
}
