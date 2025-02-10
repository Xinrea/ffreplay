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

	sprite := component.Sprite.Get(player)
	if !sprite.Initialized {
		return
	}

	status := component.Status.Get(player)

	worldM := camera.WorldMatrixInverted()

	// render target ring
	c := colorm.ColorM{}
	if global.TargetPlayer != nil && global.TargetPlayer == player {
		c.ChangeHSV(135.0/180.0*3.14, 1, 1.2)
	}

	// render tether
	tethers := sprite.Instances[0].GetTethers()
	for _, target := range tethers {
		// draw a line from player to target
		sp := sprite.Instances[0].Object.Position()
		tp := target.Object.Position()

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

	// player only has one instance
	pos := sprite.Instances[0].Object.Position()
	geoM := texture.CenterGeoM(sprite.Texture)
	geoM.Scale(sprite.Scale, sprite.Scale)
	geoM.Rotate(sprite.Instances[0].Face)
	geoM.Translate(pos[0], pos[1])
	geoM.Concat(worldM)

	if global.ShowTargetRing {
		op := &colorm.DrawImageOptions{}
		op.GeoM = geoM
		colorm.DrawImage(screen, sprite.Texture, c, op)
	}

	c = colorm.ColorM{}
	if status.IsDead() {
		c.ChangeHSV(0, 0, 1)
	}
	// render icon
	geoM = texture.CenterGeoM(status.RoleTexture())
	geoM.Scale(0.5, 0.5)
	geoM.Rotate(camera.Rotation)
	geoM.Translate(pos[0], pos[1])
	geoM.Concat(worldM)

	op := &colorm.DrawImageOptions{}
	op.GeoM = geoM
	colorm.DrawImage(screen, status.RoleTexture(), c, op)

	// render debuffs on side of player
	screenX, screenY := camera.WorldToScreen(pos[0], pos[1])
	RenderBuffList(screen, tick, status.BuffList.DeBuffs(), screenX+30/math.Pow(1.01, float64(camera.ZoomFactor)), screenY)

	// render marker on player
	if status.Marker > 0 {
		markerTexture := model.MarkerTextures[status.Marker-1]
		geoM = texture.CenterGeoM(markerTexture)
		geoM.Rotate(camera.Rotation)
		geoM.Translate(pos[0], pos[1]-30)
		geoM.Concat(worldM)
		screen.DrawImage(markerTexture, &ebiten.DrawImageOptions{GeoM: geoM})
	}
}
