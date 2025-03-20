package system

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (s *System) playgroundControl() {
	global := entry.GetGlobal()
	camera := entry.GetCamera()
	vel := vector.Vector{}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		global.TargetPlayer = nil
	}

	if global.TargetPlayer != nil {
		status := component.Status.Get(global.TargetPlayer)
		obj := status.Instances[0]

		if !status.IsDead() {
			handleRotationInput(camera)
			handleMovementInput(&vel, obj, camera)

			x, y := ebiten.CursorPosition()
			wx, wy := camera.ScreenToWorld(float64(x), float64(y))

			face := obj.Object.Position().Sub(vector.Vector{wx, wy}).Radian()
			obj.Face = util.NormalizeRadians(face + math.Pi)

			obj.Object.Translate(vel)
			// bind camera on player
			camera.Position = obj.Object.Position()
		}
	}
}

func handleMovementInput(vel *vector.Vector, obj *model.Instance, camera *model.CameraData) {
	face := camera.Rotation

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		(*vel)[1] = -MaxVelocity * math.Cos(face)
		(*vel)[0] = MaxVelocity * math.Sin(face)
	}

	if ebiten.IsKeyPressed(ebiten.KeyS) {
		(*vel)[1] = MaxVelocity * math.Cos(face)
		(*vel)[0] = -MaxVelocity * math.Sin(face)
	}

	if ebiten.IsKeyPressed(ebiten.KeyD) {
		(*vel)[1] = MaxVelocity * math.Sin(face)
		(*vel)[0] = MaxVelocity * math.Cos(face)
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		(*vel)[1] = -MaxVelocity * math.Sin(face)
		(*vel)[0] = -MaxVelocity * math.Cos(face)
	}
}

func handleRotationInput(camera *model.CameraData) {
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		camera.Rotation += 0.05
		if camera.Rotation > math.Pi {
			camera.Rotation -= 2 * math.Pi
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		camera.Rotation -= 0.05
		if camera.Rotation < -math.Pi {
			camera.Rotation += 2 * math.Pi
		}
	}
}
