package renderer

import (
	"image/color"
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/hajimehoshi/ebiten/v2"
	evector "github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi/ecs"
)

// SelectionRender draws a highlight ring around the currently selected object
// in playground mode, so the user can see what the property panel is editing.
func (r *Renderer) SelectionRender(ecs *ecs.ECS, screen *ebiten.Image) {
	global := entry.GetGlobal(ecs)
	if global.ReplayMode || global.Selected == nil {
		return
	}

	if !global.Selected.Valid() {
		global.Selected = nil

		return
	}

	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	zoom := math.Pow(1.01, float64(camera.ZoomFactor))
	highlight := color.NRGBA{255, 215, 0, 255}

	// WorldMarker: draw ring only, no facing pointer.
	if global.Selected.HasComponent(component.WorldMarker) {
		marker := component.WorldMarker.Get(global.Selected)
		sx, sy := camera.WorldToScreen(marker.Position[0], marker.Position[1])
		radius := float32(32 * zoom)
		evector.StrokeCircle(screen, float32(sx), float32(sy), radius, 2.5, highlight, true)

		return
	}

	sprite := component.Sprite.Get(global.Selected)
	if sprite == nil || global.SelectedInstance >= len(sprite.Instances) {
		return
	}

	inst := sprite.Instances[global.SelectedInstance]
	pos := inst.Object.Position()
	sx, sy := camera.WorldToScreen(pos[0], pos[1])

	// Ring radius scales with zoom so it hugs the object consistently.
	radius := float32(28 * zoom)
	evector.StrokeCircle(screen, float32(sx), float32(sy), radius, 2.5, highlight, true)

	// A short pointer line indicating facing direction.
	face := inst.Face + camera.Rotation
	fx := float32(sx) + radius*float32(math.Sin(face))
	fy := float32(sy) - radius*float32(math.Cos(face))
	evector.StrokeLine(screen, float32(sx), float32(sy), fx, fy, 2.5, highlight, true)
}
