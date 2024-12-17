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

func (s *System) MarkerUpdate(ecs *ecs.ECS) {
	if s.InReplay {
		return
	}
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			// create marker at cursor
			x, y := ebiten.CursorPosition()
			geoM := camera.WorldMatrix()
			wx, wy := geoM.Apply(float64(x), float64(y))
			entry.NewMarker(ecs, model.Marker1, f64.Vec2{wx, wy})
			log.Println("Create marker at", wx, wy)
		}
		if inpututil.IsKeyJustPressed(ebiten.Key2) {
			// create marker at cursor
			x, y := ebiten.CursorPosition()
			geoM := camera.WorldMatrix()
			wx, wy := geoM.Apply(float64(x), float64(y))
			entry.NewMarker(ecs, model.Marker2, f64.Vec2{wx, wy})
			log.Println("Create marker at", wx, wy)
		}
		if inpututil.IsKeyJustPressed(ebiten.Key3) {
			// create marker at cursor
			x, y := ebiten.CursorPosition()
			geoM := camera.WorldMatrix()
			wx, wy := geoM.Apply(float64(x), float64(y))
			entry.NewMarker(ecs, model.Marker3, f64.Vec2{wx, wy})
			log.Println("Create marker at", wx, wy)
		}
		if inpututil.IsKeyJustPressed(ebiten.Key4) {
			// create marker at cursor
			x, y := ebiten.CursorPosition()
			geoM := camera.WorldMatrix()
			wx, wy := geoM.Apply(float64(x), float64(y))
			entry.NewMarker(ecs, model.Marker4, f64.Vec2{wx, wy})
			log.Println("Create marker at", wx, wy)
		}
		if inpututil.IsKeyJustPressed(ebiten.Key5) {
			// create marker at cursor
			x, y := ebiten.CursorPosition()
			geoM := camera.WorldMatrix()
			wx, wy := geoM.Apply(float64(x), float64(y))
			entry.NewMarker(ecs, model.MarkerA, f64.Vec2{wx, wy})
			log.Println("Create marker at", wx, wy)
		}
		if inpututil.IsKeyJustPressed(ebiten.Key6) {
			// create marker at cursor
			x, y := ebiten.CursorPosition()
			geoM := camera.WorldMatrix()
			wx, wy := geoM.Apply(float64(x), float64(y))
			entry.NewMarker(ecs, model.MarkerB, f64.Vec2{wx, wy})
			log.Println("Create marker at", wx, wy)
		}
		if inpututil.IsKeyJustPressed(ebiten.Key7) {
			// create marker at cursor
			x, y := ebiten.CursorPosition()
			geoM := camera.WorldMatrix()
			wx, wy := geoM.Apply(float64(x), float64(y))
			entry.NewMarker(ecs, model.MarkerC, f64.Vec2{wx, wy})
			log.Println("Create marker at", wx, wy)
		}
		if inpututil.IsKeyJustPressed(ebiten.Key8) {
			// create marker at cursor
			x, y := ebiten.CursorPosition()
			geoM := camera.WorldMatrix()
			wx, wy := geoM.Apply(float64(x), float64(y))
			entry.NewMarker(ecs, model.MarkerD, f64.Vec2{wx, wy})
			log.Println("Create marker at", wx, wy)
		}
	}
}
