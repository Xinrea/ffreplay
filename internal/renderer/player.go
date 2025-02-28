package renderer

import (
	"image/color"
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) PlayerRender(ecs *ecs.ECS, screen *ebiten.Image) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	for e := range tag.Player.Iter(ecs.World) {
		r.renderPlayer(ecs, camera, screen, e)
	}
}

func (r *Renderer) renderPlayer(ecs *ecs.ECS, camera *model.CameraData, screen *ebiten.Image, player *donburi.Entry) {
	tick := entry.GetTick(ecs)
	global := component.Global.Get(component.Global.MustFirst(ecs.World))

	status := component.Status.Get(player)

	worldM := camera.WorldMatrixInverted()

	// render target ring
	c := colorm.ColorM{}
	if global.TargetPlayer != nil && global.TargetPlayer == player {
		c.ChangeHSV(135.0/180.0*3.14, 1, 1.2)
	}

	// render tether
	tethers := status.GetTethers()
	for _, tether := range tethers {
		// draw a line from player to target
		sp := status.Instances[0].Object.Position()
		tp := tether.Target.Instances[0].Object.Position()

		spx, spy := camera.WorldToScreen(sp[0], sp[1])
		tpx, tpy := camera.WorldToScreen(tp[0], tp[1])

		vector.StrokeLine(
			screen,
			float32(spx),
			float32(spy),
			float32(tpx),
			float32(tpy),
			4,
			color.NRGBA{255, 215, 0, 200},
			true)
	}

	status.Render(tick, camera, screen, global.ShowTargetRing, global.Debug)

	// render debuffs on side of player
	for i := range status.Instances {
		pos := status.Instances[i].Object.Position()
		screenX, screenY := camera.WorldToScreen(pos[0], pos[1])
		RenderBuffList(
			screen,
			tick,
			status.BuffList.DeBuffs(),
			screenX+30/math.Pow(1.01, float64(camera.ZoomFactor)),
			screenY)

		// render marker on player
		if status.Marker > 0 {
			markerTexture := model.MarkerTextures[status.Marker-1]
			geoM := texture.CenterGeoM(markerTexture)
			geoM.Rotate(camera.Rotation)
			geoM.Translate(pos[0], pos[1]-30)
			geoM.Concat(worldM)
			screen.DrawImage(markerTexture, &ebiten.DrawImageOptions{GeoM: geoM})
		}

		break
	}
}
