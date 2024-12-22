package renderer

import (
	"log"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) BackgroundRender(ecs *ecs.ECS, screen *ebiten.Image) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	for g := range tag.Background.Iter(ecs.World) {
		ground := component.Map.Get(g)
		if ground.Config.Phases == nil {
			r.mapRender(camera, screen, ground.Config.DefaultMap)
			continue
		}
		// find current phase
		p := entry.GetPhase(ecs)
		if p < 0 || p >= len(ground.Config.Phases) {
			log.Println("phase not found", p)
			p = 0
		}
		r.mapRender(camera, screen, ground.Config.Phases[p])
	}
}

func (r *Renderer) mapRender(camera *model.CameraData, screen *ebiten.Image, m model.MapItem) {
	geoM := m.Texture.GetGeoM()
	if m.Scale > 0 {
		geoM.Scale(m.Scale, m.Scale)
	}
	geoM.Translate(m.Offset.X, m.Offset.Y)
	wordM := camera.WorldMatrix()
	wordM.Invert()
	geoM.Concat(wordM)
	screen.DrawImage(m.Texture.Img(), &ebiten.DrawImageOptions{
		Filter: ebiten.FilterLinear,
		GeoM:   geoM,
	})
}
