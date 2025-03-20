package system

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/math/f64"
)

func (s *System) WorldMarkerUpdate() {
	global := entry.GetGlobal()
	if global.ReplayMode {
		return
	}

	camera := entry.GetCamera()

	if global.WorldMarkerSelected >= 0 {
		if !global.UIFocus && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			global.WorldMarkerSelected = -1
		}
		// Update selected marker
		worldMarkers := entry.GetWorldMarkers()
		for _, marker := range worldMarkers {
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
