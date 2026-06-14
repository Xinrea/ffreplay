package system

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/yohamta/donburi/ecs"
)

// WalkUpdate handles walk-to animation for entities.
// Each frame, entities with an active WalkTarget are interpolated toward their destination.
func (s *System) WalkUpdate(ecs *ecs.ECS) {
	tick := entry.GetTick(ecs)

	for e := range tag.GameObject.Iter(ecs.World) {
		sprite := component.Sprite.Get(e)
		if sprite == nil {
			continue
		}

		for _, inst := range sprite.Instances {
			if inst.WalkTarget == nil {
				continue
			}

			// Initialize walk start on first frame
			if inst.WalkTick < 0 {
				inst.WalkTick = tick
				inst.WalkStart = inst.Object.Position()

				// Face the movement direction. Face=0 points north (up),
				// and forward is (sin(Face), -cos(Face)), so for a movement
				// vector (dx, dy): Face = atan2(dx, -dy).
				dst := inst.WalkTarget.Position()
				dx := dst[0] - inst.WalkStart[0]
				dy := dst[1] - inst.WalkStart[1]
				if dx != 0 || dy != 0 {
					inst.Face = math.Atan2(dx, -dy)
				}
			}

			elapsed := tick - inst.WalkTick
			if elapsed >= inst.WalkDuration {
				// Destination reached
				inst.Object.UpdatePosition(inst.WalkTarget.Position())
				inst.WalkTarget = nil
				inst.WalkTick = 0
				inst.WalkDuration = 0

				continue
			}

			// Linear interpolation
			t := float64(elapsed) / float64(inst.WalkDuration)
			// Ease-in-out for smoother movement
			t = easeInOutQuad(t)

			start := inst.WalkStart
			end := inst.WalkTarget.Position()
			currentX := start[0] + (end[0]-start[0])*t
			currentY := start[1] + (end[1]-start[1])*t
			inst.Object.UpdatePosition(vector.NewVector(currentX, currentY))
		}
	}
}

// easeInOutQuad provides a smooth ease-in-out curve.
func easeInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}
