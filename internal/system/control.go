package system

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

const MaxVelocity = 8

func (s *System) ControlUpdate(ecs *ecs.ECS) {
	globalData := component.Global.Get(tag.Global.MustFirst(ecs.World))
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	camera.Update(s.ViewPort)

	vel := vector.Vector{}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackquote) {
		globalData.Debug = !globalData.Debug
	}

	if !s.InReplay {
		var player *donburi.Entry = tag.Player.MustFirst(ecs.World)
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
			if !s.InReplay {
				camera.Rotation = obj.Face
			}
		}
	}

	if s.InReplay {
		// replaying control
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			s.Pause = !s.Pause
			globalData.Speed = 10
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			if globalData.Speed == 5 {
				globalData.Speed = 10
				return
			}
			globalData.Speed = min(globalData.Speed+10, 50)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			if globalData.Speed > 10 {
				globalData.Speed -= 10
			} else {
				globalData.Speed = 5
			}
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			globalData.Tick -= 60 * 10 // 1s tick
			globalData.Tick = max(0, globalData.Tick)
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			globalData.Tick += 60 * 10 // 1s tick
			globalData.Tick = min(globalData.Tick, util.MSToTick(globalData.FightDuration.Load())*10)
		}

		// lock view of player
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			s.TargetPlayer = 1
		}
		if inpututil.IsKeyJustPressed(ebiten.Key2) {
			s.TargetPlayer = 2
		}
		if inpututil.IsKeyJustPressed(ebiten.Key3) {
			s.TargetPlayer = 3
		}
		if inpututil.IsKeyJustPressed(ebiten.Key4) {
			s.TargetPlayer = 4
		}
		if inpututil.IsKeyJustPressed(ebiten.Key5) {
			s.TargetPlayer = 5
		}
		if inpututil.IsKeyJustPressed(ebiten.Key6) {
			s.TargetPlayer = 6
		}
		if inpututil.IsKeyJustPressed(ebiten.Key7) {
			s.TargetPlayer = 7
		}
		if inpututil.IsKeyJustPressed(ebiten.Key8) {
			s.TargetPlayer = 8
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			s.TargetPlayer = 0
		}

		// camera control
		if ebiten.IsKeyPressed(ebiten.KeyE) {
			camera.Rotation += 0.05
		}
		if ebiten.IsKeyPressed(ebiten.KeyQ) {
			camera.Rotation -= 0.05
		}

		var targetPlayer *donburi.Entry = nil
		if s.TargetPlayer != 0 && s.TargetPlayer-1 < len(s.PlayerList) {
			targetPlayer = s.PlayerList[s.TargetPlayer-1]
		}
		// if view not locked
		if targetPlayer == nil {
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
			camera.Position = component.Sprite.Get(targetPlayer).Instances[0].Object.Position()
		}
	}

	_, dy := ebiten.Wheel()
	if util.IsWasm() {
		camera.ZoomFactor -= int(dy)
	} else {
		camera.ZoomFactor -= int(dy * 3)
	}
}
