package renderer

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) WorldMarkerRender(ecs *ecs.ECS, screen *ebiten.Image) {
	global := entry.GetGlobal()
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))

	if !global.Loaded.Load() {
		return
	}

	for e := range tag.WorldMarker.Iter(ecs.World) {
		marker := component.WorldMarker.Get(e)
		config := model.WorldMarkerConfigs[marker.Type]

		geoM := texture.CenterGeoM(config.Texture)
		geoM.Scale(0.5, 0.5)

		if marker.Type <= model.WorldMarkerD {
			geoM.Scale(1.1, 1.1)
		}

		geoM.Translate(marker.Position[0], marker.Position[1])
		geoM.Concat(camera.WorldMatrixInverted())
		screen.DrawImage(config.Background, &ebiten.DrawImageOptions{GeoM: geoM})

		geoM = texture.CenterGeoM(config.Texture)
		geoM.Scale(0.5, 0.5)
		geoM.Rotate(camera.Rotation)
		geoM.Translate(marker.Position[0], marker.Position[1])
		geoM.Concat(camera.WorldMatrixInverted())
		screen.DrawImage(config.Texture, &ebiten.DrawImageOptions{
			GeoM: geoM,
		})
	}
}
