package renderer

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) WorldMarkerRender(ecs *ecs.ECS, screen *ebiten.Image) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	for e := range tag.WorldMarker.Iter(ecs.World) {
		marker := component.WorldMarker.Get(e)
		config := model.WorldMarkerConfigs[marker.Type]
		geoM := config.Texture.GetGeoM()
		geoM.Scale(0.5, 0.5)
		if marker.Type <= model.WorldMarkerD {
			geoM.Scale(1.1, 1.1)
		}
		geoM.Translate(marker.Position[0], marker.Position[1])
		geoM.Concat(camera.WorldMatrixInverted())
		screen.DrawImage(config.Background, &ebiten.DrawImageOptions{GeoM: geoM})

		geoM = config.Texture.GetGeoM()
		geoM.Scale(0.5, 0.5)
		geoM.Rotate(camera.Rotation)
		geoM.Translate(marker.Position[0], marker.Position[1])
		geoM.Concat(camera.WorldMatrixInverted())
		screen.DrawImage(config.Texture.Img(), &ebiten.DrawImageOptions{
			GeoM: geoM,
		})
	}
}
