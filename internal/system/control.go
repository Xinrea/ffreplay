package system

import (
	"log"
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

	var player *donburi.Entry = nil
	for e := range tag.Player.Iter(ecs.World) {
		if component.Status.Get(e).Role == s.MainPlayerRole {
			player = e
			break
		}
	}
	if player == nil {
		log.Fatal("Player not found")
	}
	status := component.Status.Get(player)
	obj := component.Sprite.Get(player)
	vel := component.Velocity.Get(player)

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.Pause = !s.Pause
		globalData.Speed = 10
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		globalData.Speed = min(globalData.Speed+10, 50)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if globalData.Speed > 10 {
			globalData.Speed -= 10
		} else {
			globalData.Speed = 5
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		globalData.Tick -= 60 * 10 * 10 // 10s tick
		globalData.Tick = max(0, globalData.Tick)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		globalData.Tick += 60 * 10 * 10 // 10s tick
	}
	if !status.IsDead() && !s.InReplay {
		// remember that face is relative to north
		vel.Clear()
		if ebiten.IsKeyPressed(ebiten.KeyW) {
			vel.Y = -MaxVelocity * math.Cos(obj.Face)
			vel.X = MaxVelocity * math.Sin(obj.Face)
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) {
			vel.Y = MaxVelocity * math.Cos(obj.Face)
			vel.X = -MaxVelocity * math.Sin(obj.Face)
		}
		if ebiten.IsKeyPressed(ebiten.KeyD) {
			vel.Y = MaxVelocity * math.Sin(obj.Face)
			vel.X = MaxVelocity * math.Cos(obj.Face)
		}
		if ebiten.IsKeyPressed(ebiten.KeyA) {
			vel.Y = -MaxVelocity * math.Sin(obj.Face)
			vel.X = -MaxVelocity * math.Cos(obj.Face)
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
		obj.Object.Translate(vector.Vector{vel.X, vel.Y})
		// bind camera on player
		camera.Position = obj.Object.Position()
		if !s.InReplay {
			camera.Rotation = obj.Face
		}
	}

	if s.InReplay {
		vel.Clear()
		if ebiten.IsKeyPressed(ebiten.KeyE) {
			camera.Rotation += 0.05
		}
		if ebiten.IsKeyPressed(ebiten.KeyQ) {
			camera.Rotation -= 0.05
		}
		if ebiten.IsKeyPressed(ebiten.KeyW) {
			vel.Y = -MaxVelocity * math.Cos(camera.Rotation) * 2
			vel.X = MaxVelocity * math.Sin(camera.Rotation) * 2
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) {
			vel.Y = MaxVelocity * math.Cos(camera.Rotation) * 2
			vel.X = -MaxVelocity * math.Sin(camera.Rotation) * 2
		}
		if ebiten.IsKeyPressed(ebiten.KeyD) {
			vel.Y = MaxVelocity * math.Sin(camera.Rotation) * 2
			vel.X = MaxVelocity * math.Cos(camera.Rotation) * 2
		}
		if ebiten.IsKeyPressed(ebiten.KeyA) {
			vel.Y = -MaxVelocity * math.Sin(camera.Rotation) * 2
			vel.X = -MaxVelocity * math.Cos(camera.Rotation) * 2
		}
		camera.Position = camera.Position.Add(vector.Vector{vel.X, vel.Y}.Scale(math.Pow(1.01, float64(camera.ZoomFactor))))
	}

	_, dy := ebiten.Wheel()
	if util.IsWasm() {
		camera.ZoomFactor -= int(dy)
	} else {
		camera.ZoomFactor -= int(dy * 3)
	}
}
