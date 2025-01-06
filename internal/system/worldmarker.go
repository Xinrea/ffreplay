package system

import (
	"log"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/image/math/f64"
)

func (s *System) WorldMarkerUpdate(ecs *ecs.ECS) {
	if entry.GetGlobal(s.ecs).ReplayMode {
		return
	}

	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))

	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		for key, marker := range map[ebiten.Key]model.WorldMarkerType{
			ebiten.Key1: model.WorldMarker1,
			ebiten.Key2: model.WorldMarker2,
			ebiten.Key3: model.WorldMarker3,
			ebiten.Key4: model.WorldMarker4,
			ebiten.Key5: model.WorldMarkerA,
			ebiten.Key6: model.WorldMarkerB,
			ebiten.Key7: model.WorldMarkerC,
			ebiten.Key8: model.WorldMarkerD,
		} {
			if inpututil.IsKeyJustPressed(key) {
				createMarkerAtCursor(ecs, camera, marker)
			}
		}
	}
}

func createMarkerAtCursor(ecs *ecs.ECS, camera *model.CameraData, marker model.WorldMarkerType) {
	x, y := ebiten.CursorPosition()
	wx, wy := camera.ScreenToWorld(float64(x), float64(y))
	entry.NewWorldMarker(ecs, marker, f64.Vec2{wx, wy})
	log.Println("Create marker at", wx, wy)
}
