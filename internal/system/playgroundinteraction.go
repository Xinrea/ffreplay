package system

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// dragMode describes the current pointer drag interaction in the playground.
type dragMode int

const (
	dragNone dragMode = iota
	// dragPan: holding Space + left mouse drags the camera (grab).
	dragPan
	// dragObject: dragging a selected object to a new position.
	dragObject
)

// pickRadius is the screen-space radius (in pixels) used to hit-test clicks
// against object positions.
const pickRadius = 40.0

// PlaygroundInteractionUpdate handles mouse-driven interaction in the
// playground: Space-hold grab panning, click-to-select, and drag-to-move.
// It is only active when not in replay mode.
func (s *System) PlaygroundInteractionUpdate(ecs *ecs.ECS) {
	global := entry.GetGlobal(ecs)
	if global.ReplayMode || !global.Loaded.Load() {
		return
	}

	// Never interact through focused inputs or hovered UI panels.
	if global.UIFocus || global.UIHovered {
		s.endDrag()

		return
	}

	camera := entry.GetCamera(ecs)
	mx, my := ebiten.CursorPosition()

	spaceHeld := ebiten.IsKeyPressed(ebiten.KeySpace)
	s.updateCursorShape(spaceHeld)

	switch s.dragMode {
	case dragNone:
		s.handleDragStart(ecs, global, camera, mx, my, spaceHeld)
	case dragPan:
		s.handlePanDrag(camera, mx, my)
	case dragObject:
		s.handleObjectDrag(global, camera, mx, my)
	}

	// Any release ends the active drag.
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		s.endDrag()
	}
}

// updateCursorShape reflects the grab affordance while Space is held or a drag
// is in progress.
func (s *System) updateCursorShape(spaceHeld bool) {
	if spaceHeld || s.dragMode != dragNone {
		ebiten.SetCursorShape(ebiten.CursorShapeMove)

		return
	}

	// Only reset when we previously owned the grab/move shape, so UI handlers
	// (e.g. the input box's text cursor) keep control of their own shape.
	if ebiten.CursorShape() == ebiten.CursorShapeMove {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}
}

// handleDragStart decides what a fresh left-press begins: panning the camera
// (when Space is held) or selecting/dragging an object.
func (s *System) handleDragStart(
	ecs *ecs.ECS,
	global *model.GlobalData,
	camera *model.CameraData,
	mx, my int,
	spaceHeld bool,
) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	if spaceHeld {
		s.dragMode = dragPan
		s.lastMouseX = mx
		s.lastMouseY = my

		return
	}

	// Hit-test an object under the cursor (in screen space).
	hitEntry, hitInstance := s.pickEntity(ecs, camera, float64(mx), float64(my))

	if hitEntry != nil {
		global.Selected = hitEntry
		global.SelectedInstance = hitInstance
		s.dragMode = dragObject
		s.lastMouseX = mx
		s.lastMouseY = my

		return
	}

	// Clicking empty space clears the selection.
	global.Selected = nil
	global.SelectedInstance = 0
}

// handlePanDrag moves the camera so the world point under the cursor stays
// fixed as the mouse moves (grab feel).
func (s *System) handlePanDrag(camera *model.CameraData, mx, my int) {
	// Map both the previous and current cursor positions into world space.
	// The difference cancels the camera-position term of the transform, so it
	// yields the exact world delta regardless of zoom/rotation/device scale.
	curX, curY := camera.ScreenToWorld(float64(mx), float64(my))
	lastX, lastY := camera.ScreenToWorld(float64(s.lastMouseX), float64(s.lastMouseY))

	s.lastMouseX = mx
	s.lastMouseY = my

	camera.Position = camera.Position.Add(vector.NewVector(lastX-curX, lastY-curY))
}

// handleObjectDrag moves the selected instance to follow the cursor.
func (s *System) handleObjectDrag(global *model.GlobalData, camera *model.CameraData, mx, my int) {
	if global.Selected == nil {
		s.dragMode = dragNone

		return
	}

	wx, wy := camera.ScreenToWorld(float64(mx), float64(my))

	// WorldMarker: update Position directly.
	if markerData := component.WorldMarker.Get(global.Selected); markerData != nil {
		markerData.Position[0] = wx
		markerData.Position[1] = wy

		return
	}

	sprite := component.Sprite.Get(global.Selected)
	if sprite == nil || global.SelectedInstance >= len(sprite.Instances) {
		s.dragMode = dragNone

		return
	}

	inst := sprite.Instances[global.SelectedInstance]
	inst.Object.UpdatePosition(vector.NewVector(wx, wy))
	// Cancel any in-flight walk animation so the manual drag wins.
	inst.WalkTarget = nil
	inst.WalkTick = 0
	inst.WalkDuration = 0
}

// endDrag clears the active drag and resets the grab cursor.
func (s *System) endDrag() {
	if s.dragMode != dragNone {
		s.dragMode = dragNone
		if ebiten.CursorShape() == ebiten.CursorShapeMove {
			ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		}
	}
}

// pickEntity returns the closest selectable game object (player or enemy)
// whose on-screen position is within pickRadius screen pixels of the cursor,
// or nil if none is close enough. Hit-testing in screen space keeps the pick
// tolerance constant regardless of zoom.
func (s *System) pickEntity(
	ecs *ecs.ECS,
	camera *model.CameraData,
	screenX, screenY float64,
) (*donburi.Entry, int) {
	// In playground mode all entities are always "active" — there is no
	// event timeline to bound their lifespan. We only need to check that
	// the sprite is initialized before hit-testing its instances.
	var (
		best     *donburi.Entry
		bestInst int
		bestDist = pickRadius
	)

	consider := func(e *donburi.Entry) {
		sprite := component.Sprite.Get(e)
		if sprite == nil || !sprite.Initialized {
			return
		}

		for i, inst := range sprite.Instances {
			pos := inst.Object.Position()
			sx, sy := camera.WorldToScreen(pos[0], pos[1])
			d := math.Hypot(sx-screenX, sy-screenY)

			if d <= bestDist {
				best = e
				bestInst = i
				bestDist = d
			}
		}
	}

	for e := range tag.Player.Iter(ecs.World) {
		consider(e)
	}

	for e := range tag.Enemy.Iter(ecs.World) {
		consider(e)
	}

	// Also hit-test WorldMarker entities.
	for e := range tag.WorldMarker.Iter(ecs.World) {
		marker := component.WorldMarker.Get(e)
		sx, sy := camera.WorldToScreen(marker.Position[0], marker.Position[1])
		d := math.Hypot(sx-screenX, sy-screenY)

		if d <= bestDist {
			best = e
			bestInst = 0
			bestDist = d
		}
	}

	return best, bestInst
}
