package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/image/math/f64"
)

func (s *System) WorldMarkerUpdate(ecs *ecs.ECS) {
	global := entry.GetGlobal(s.ecs)
	if global.ReplayMode {
		return
	}

	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))

	if global.WorldMarkerSelected >= 0 {
		if !global.UIFocus && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			global.WorldMarkerSelected = -1
		}
		// Update selected marker
		for marker := range component.WorldMarker.Iter(ecs.World) {
			markerData := component.WorldMarker.Get(marker)
			if int(markerData.Type) == global.WorldMarkerSelected {
				x, y := ebiten.CursorPosition()
				wx, wy := camera.ScreenToWorld(float64(x), float64(y))
				markerData.Position = f64.Vec2{wx, wy}

				break
			}
		}
	}
}
