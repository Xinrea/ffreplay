package system

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi/ecs"
)

func (s *System) playgroundControl(ecs *ecs.ECS) {
	global := entry.GetGlobal(ecs)
	camera := entry.GetCamera(ecs)
	vel := vector.Vector{}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		global.TargetPlayer = nil
	}

	if global.TargetPlayer != nil {
		obj := component.Sprite.Get(global.TargetPlayer).Instances[0]
		status := component.Status.Get(global.TargetPlayer)

		if !status.IsDead() {
			handleMovementInput(&vel, obj)
			handleRotationInput(obj)
			obj.Object.Translate(vel)
			// bind camera on player
			camera.Position = obj.Object.Position()
			camera.Rotation = obj.Face
		}
	}
}

func handleMovementInput(vel *vector.Vector, obj *model.Instance) {
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		(*vel)[1] = -MaxVelocity * math.Cos(obj.Face)
		(*vel)[0] = MaxVelocity * math.Sin(obj.Face)
	}

	if ebiten.IsKeyPressed(ebiten.KeyS) {
		(*vel)[1] = MaxVelocity * math.Cos(obj.Face)
		(*vel)[0] = -MaxVelocity * math.Sin(obj.Face)
	}

	if ebiten.IsKeyPressed(ebiten.KeyD) {
		(*vel)[1] = MaxVelocity * math.Sin(obj.Face)
		(*vel)[0] = MaxVelocity * math.Cos(obj.Face)
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		(*vel)[1] = -MaxVelocity * math.Sin(obj.Face)
		(*vel)[0] = -MaxVelocity * math.Cos(obj.Face)
	}
}

func handleRotationInput(obj *model.Instance) {
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
}
