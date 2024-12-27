package system

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi/ecs"
)

const MaxVelocity = 8

func (s *System) ControlUpdate(ecs *ecs.ECS) {
	global := entry.GetGlobal(s.ecs)
	camera := entry.GetCamera(s.ecs)
	camera.Update(s.ViewPort)

	if inpututil.IsKeyJustPressed(ebiten.KeyBackquote) {
		global.Debug = !global.Debug
	}
	_, dy := ebiten.Wheel()
	if util.IsWasm() {
		camera.ZoomFactor -= int(dy)
	} else {
		camera.ZoomFactor -= int(dy * 3)
	}

	if global.ReplayMode {
		s.replayModeControl(ecs)
	} else {
		s.playgroundControl(ecs)
	}
}

func (s *System) playgroundControl(ecs *ecs.ECS) {
	camera := entry.GetCamera(ecs)
	vel := vector.Vector{}
	if player, ok := tag.Player.First(ecs.World); ok {
		status := component.Status.Get(player)
		obj := component.Sprite.Get(player).Instances[0]
		if !status.IsDead() {
			// remember that face is relative to north
			if ebiten.IsKeyPressed(ebiten.KeyW) {
				vel[1] = -MaxVelocity * math.Cos(obj.Face)
				vel[0] = MaxVelocity * math.Sin(obj.Face)
			}
			if ebiten.IsKeyPressed(ebiten.KeyS) {
				vel[1] = MaxVelocity * math.Cos(obj.Face)
				vel[0] = -MaxVelocity * math.Sin(obj.Face)
			}
			if ebiten.IsKeyPressed(ebiten.KeyD) {
				vel[1] = MaxVelocity * math.Sin(obj.Face)
				vel[0] = MaxVelocity * math.Cos(obj.Face)
			}
			if ebiten.IsKeyPressed(ebiten.KeyA) {
				vel[1] = -MaxVelocity * math.Sin(obj.Face)
				vel[0] = -MaxVelocity * math.Cos(obj.Face)
			}

			if ebiten.IsKeyPressed(ebiten.KeyE) {
				obj.Face += 0.05
				if obj.Face > math.Pi {
					obj.Face -= 2 * math.Pi
				}
			}
			if ebiten.IsKeyPressed(ebiten.KeyQ) {
				obj.Face -= 0.05
				if obj.Face < -math.Pi {
					obj.Face += 2 * math.Pi
				}
			}
			obj.Object.Translate(vel)
			// bind camera on player
			camera.Position = obj.Object.Position()
			camera.Rotation = obj.Face
		}
	}
}

func (s *System) replayModeControl(ecs *ecs.ECS) {
	global := entry.GetGlobal(ecs)
	camera := entry.GetCamera(ecs)
	vel := vector.Vector{}
	// replaying control
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
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		global.Tick -= 60 * 10 // 1s tick
		global.Tick = max(0, global.Tick)
		s.doReset(ecs)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		global.Tick += 60 * 10 // 1s tick
		global.Tick = min(global.Tick, util.MSToTick(global.FightDuration.Load())*10)
	}

	// lock view of player
	newTarget := -1
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		newTarget = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		newTarget = 2
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		newTarget = 3
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		newTarget = 4
	}
	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		newTarget = 5
	}
	if inpututil.IsKeyJustPressed(ebiten.Key6) {
		newTarget = 6
	}
	if inpututil.IsKeyJustPressed(ebiten.Key7) {
		newTarget = 7
	}
	if inpututil.IsKeyJustPressed(ebiten.Key8) {
		newTarget = 8
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		newTarget = 0
	}

	// did selection
	if newTarget != -1 {
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

	// camera control
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		camera.Rotation += 0.05
	}
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		camera.Rotation -= 0.05
	}

	// if view not locked
	if global.TargetPlayer == nil {
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
	} else {
		// bind camera on target player
		camera.Position = component.Sprite.Get(global.TargetPlayer).Instances[0].Object.Position()
	}

}
