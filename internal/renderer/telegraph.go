package renderer

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/Xinrea/ffreplay/pkg/vector"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
)

// TelegraphRender renders temporary visual telegraphs (AoE indicators, text annotations).
func (r *Renderer) TelegraphRender(ecs *ecs.ECS, screen *ebiten.Image) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	worldM := camera.WorldMatrixInverted()

	for e := range tag.Telegraph.Iter(ecs.World) {
		td := component.Telegraph.Get(e)
		if td == nil {
			continue
		}

		// Calculate alpha fade based on remaining duration
		alpha := td.Color.A
		if td.MaxDuration > 0 && td.Duration > 0 {
			ratio := float64(td.Duration) / float64(td.MaxDuration)
			alpha = uint8(float64(alpha) * math.Min(ratio*2, 1.0))
		}

		switch td.Type {
		case model.TelegraphCircle, model.TelegraphRect:
			c := td.Color
			c.A = alpha
			td.Object.Render(screen, worldM, c)

		case model.TelegraphText:
			pos := td.Object.Position()
			sx, sy := camera.WorldToScreen(pos[0], pos[1])
			s := ebiten.Monitor().DeviceScaleFactor()
			c := td.Color
			c.A = alpha
			ui.DrawText(
				screen,
				td.Text,
				16*s,
				sx,
				sy,
				c,
				furex.AlignItemCenter,
				textShdowOpt,
			)
		}
	}
}

// HeadMarkerRender renders head marker icons above entities that have them.
func (r *Renderer) HeadMarkerRender(ecs *ecs.ECS, screen *ebiten.Image) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	worldM := camera.WorldMatrixInverted()

	for e := range component.Status.Iter(ecs.World) {
		status := component.Status.Get(e)
		if status.HeadMarker <= 0 {
			continue
		}

		sprite := component.Sprite.Get(e)
		if sprite == nil || len(sprite.Instances) == 0 {
			continue
		}

		pos := sprite.Instances[0].Object.Position()

		// Render marker above entity
		markerTex := model.GetHeadMarkerTexture(status.HeadMarker)
		if markerTex == nil {
			continue
		}

		// Offset above the entity head
		pos = pos.Add(vector.Vector{0, -35})

		geoM := texture.CenterGeoM(markerTex)
		geoM.Scale(0.5, 0.5)
		geoM.Translate(pos[0], pos[1])
		geoM.Concat(worldM)

		cm := colorm.ColorM{}
		colorm.DrawImage(screen, markerTex, cm, &colorm.DrawImageOptions{
			GeoM:   geoM,
			Filter: ebiten.FilterLinear,
		})
	}
}
