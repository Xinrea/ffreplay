package renderer

import (
	"log"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) BackgroundRender(ecs *ecs.ECS, screen *ebiten.Image) {
	global := component.Global.Get(tag.Global.MustFirst(ecs.World))
	if !global.Loaded.Load() {
		return
	}
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	g, ok := tag.Background.First(ecs.World)
	if !ok {
		return
	}
	ground := component.Map.Get(g)
	if len(ground.Config.Phases) > 0 {
		// find current phase
		p := entry.GetPhase(ecs)
		if p < 0 || p >= len(ground.Config.Phases) {
			p = 0
		}
		r.mapRender(camera, screen, ground.Config.Phases[p])
		return
	}
	// have no choice
	if len(ground.Config.Maps) == 1 {
		for _, m := range ground.Config.Maps {
			r.mapRender(camera, screen, m)
			return
		}
	}
	if m, ok := ground.Config.Maps[ground.Config.CurrentMap]; ok {
		r.mapRender(camera, screen, m)
	} else {
		log.Println("Current map not found")
	}
}

func (r *Renderer) mapRender(camera *model.CameraData, screen *ebiten.Image, m model.MapItem) {
	geoM := texture.CenterGeoM(m.Texture)
	if m.Scale > 0 {
		geoM.Scale(m.Scale, m.Scale)
	}
	geoM.Translate(m.Offset.X*25, m.Offset.Y*25)
	wordM := camera.WorldMatrixInverted()
	geoM.Concat(wordM)
	screen.DrawImage(m.Texture, &ebiten.DrawImageOptions{
		Filter: ebiten.FilterLinear,
		GeoM:   geoM,
	})
}
