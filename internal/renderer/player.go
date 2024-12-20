package renderer

import (
	"math"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
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
	wordM := camera.WorldMatrix()
	wordM.Invert()

	// render target ring
	c := colorm.ColorM{}
	if global.TargetPlayer != nil && global.TargetPlayer == player {
		c.ChangeHSV(135.0/180.0*3.14, 1, 1.2)
	}
	// player only has one instance
	pos := sprite.Instances[0].Object.Position()
	geoM := sprite.Texture.GetGeoM()
	geoM.Scale(sprite.Scale, sprite.Scale)
	geoM.Rotate(sprite.Instances[0].Face)
	geoM.Translate(pos[0], pos[1])
	geoM.Concat(wordM)
	op := &colorm.DrawImageOptions{}
	op.GeoM = geoM
	colorm.DrawImage(screen, sprite.Texture.Img(), c, op)

	c = colorm.ColorM{}
	if status.IsDead() {
		c.ChangeHSV(0, 0, 1)
	}
	// render icon
	geoM = status.RoleTexture().GetGeoM()
	geoM.Scale(0.5, 0.5)
	geoM.Rotate(camera.Rotation)
	geoM.Translate(pos[0], pos[1])
	geoM.Concat(wordM)
	op = &colorm.DrawImageOptions{}
	op.GeoM = geoM
	colorm.DrawImage(screen, status.RoleTexture().Img(), c, op)

	// render debuffs on side of player
	s := ebiten.Monitor().DeviceScaleFactor()
	screenX, screenY := camera.WorldToScreen(pos[0], pos[1])
	RenderBuffList(screen, tick, status.BuffList.DeBuffs(), screenX/s+30/math.Pow(1.01, float64(camera.ZoomFactor)), screenY/s, ebiten.Monitor().DeviceScaleFactor())
}
