package system

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (s *System) replayModeControl() {
	global := entry.GetGlobal()
	camera := entry.GetCamera()

	s.handleSpeedControl()
	s.handleRewindControl()
	s.handleTargetSelection()

	if ebiten.IsKeyPressed(ebiten.KeyE) {
		camera.Rotation += 0.05
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		camera.Rotation -= 0.05
	}

	if global.TargetPlayer != nil {
		camera.Position = component.Status.Get(global.TargetPlayer).Instances[0].Object.Position()

		return
	}

	s.handleCameraPosControl()
}

func (s *System) handleSpeedControl() {
	global := entry.GetGlobal()

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.Pause = !s.Pause
		global.Speed = 10
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		if global.Speed == 5 {
			global.Speed = 10

			return
		}

		global.Speed = min(global.Speed+10, 50)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if global.Speed > 10 {
			global.Speed -= 10
		} else {
			global.Speed = 5
		}
	}
}

func (s *System) handleRewindControl() {
	global := entry.GetGlobal()

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		global.Tick -= 60 * 10 // 1s tick
		global.Tick = max(0, global.Tick)
		global.Reset.Store(true)
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		global.Tick += 60 * 10 // 1s tick
		global.Tick = min(global.Tick, util.MSToTick(global.FightDuration.Load())*10)
	}
}

func (s *System) handleTargetSelection() {
	newTarget := -1

	for i := ebiten.Key1; i <= ebiten.Key8; i++ {
		if inpututil.IsKeyJustPressed(i) {
			newTarget = int(i - ebiten.Key1 + 1)

			break
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		newTarget = 0
	}

	if newTarget != -1 {
		s.setTargetPlayer(newTarget)
	}
}

func (s *System) setTargetPlayer(newTarget int) {
	global := entry.GetGlobal()
	if newTarget == 0 {
		global.TargetPlayer = nil
	} else {
		newTarget -= 1
		if newTarget < len(s.PlayerList) {
			global.TargetPlayer = s.PlayerList[newTarget]
		} else {
			global.TargetPlayer = nil
		}
	}
}

func (s *System) handleCameraPosControl() {
	camera := entry.GetCamera()
	vel := vector.Vector{}

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		vel[1] = -MaxVelocity * math.Cos(camera.Rotation) * 2
		vel[0] = MaxVelocity * math.Sin(camera.Rotation) * 2
	}

	if ebiten.IsKeyPressed(ebiten.KeyS) {
		vel[1] = MaxVelocity * math.Cos(camera.Rotation) * 2
		vel[0] = -MaxVelocity * math.Sin(camera.Rotation) * 2
	}

	if ebiten.IsKeyPressed(ebiten.KeyD) {
		vel[1] = MaxVelocity * math.Sin(camera.Rotation) * 2
		vel[0] = MaxVelocity * math.Cos(camera.Rotation) * 2
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		vel[1] = -MaxVelocity * math.Sin(camera.Rotation) * 2
		vel[0] = -MaxVelocity * math.Cos(camera.Rotation) * 2
	}

	camera.Position = camera.Position.Add(vel.Scale(math.Pow(1.01, float64(camera.ZoomFactor))))
}
