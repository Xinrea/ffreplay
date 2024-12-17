package renderer

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) EnemyRender(ecs *ecs.ECS, screen *ebiten.Image) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	for e := range tag.Enemy.Iter(ecs.World) {
		r.renderEnemy(entry.GetTick(ecs), camera, screen, e)
	}
}

func (r *Renderer) renderEnemy(tick int64, camera *model.CameraData, screen *ebiten.Image, player *donburi.Entry) {
	sprite := component.Sprite.Get(player)
	status := component.Status.Get(player)
	if sprite.Texture == nil {
		return
	}
	if util.TickToMS(tick-status.LastActive) > 5000 {
		return
	}
	wordM := camera.WorldMatrix()
	wordM.Invert()

	var c colorm.ColorM
	// render target ring
	geoM := sprite.Texture.GetGeoM()
	if status.Role == model.NPC {
		geoM.Scale(0.5, 0.5)
	}
	geoM.Rotate(sprite.Face)
	geoM.Translate(sprite.Object.Position()[0], sprite.Object.Position()[1])
	geoM.Concat(wordM)
	op := &colorm.DrawImageOptions{}
	op.GeoM = geoM
	colorm.DrawImage(screen, sprite.Texture.Img(), c, op)
}
