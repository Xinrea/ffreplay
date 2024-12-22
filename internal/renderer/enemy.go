package renderer

import (
	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) EnemyRender(ecs *ecs.ECS, screen *ebiten.Image) {
	for e := range tag.Enemy.Iter(ecs.World) {
		r.renderEnemy(ecs, screen, e)
	}
}

func (r *Renderer) renderEnemy(ecs *ecs.ECS, screen *ebiten.Image, enemy *donburi.Entry) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	global := component.Global.Get(tag.Global.MustFirst(ecs.World))
	tick := entry.GetTick(ecs)
	sprite := component.Sprite.Get(enemy)
	if !sprite.Initialized {
		return
	}
	status := component.Status.Get(enemy)
	renderObject := func(face float64, obj object.Object) {
		pos := obj.Position()
		if sprite.Texture == nil {
			return
		}
		wordM := camera.WorldMatrix()
		wordM.Invert()

		var c colorm.ColorM
		// render target ring
		if global.Debug || status.Role == model.Boss {
			geoM := sprite.Texture.GetGeoM()
			if status.Role == model.NPC {
				geoM.Scale(0.5, 0.5)
			}
			geoM.Rotate(face)
			geoM.Translate(pos[0], pos[1])
			geoM.Concat(wordM)
			op := &colorm.DrawImageOptions{}
			op.GeoM = geoM
			colorm.DrawImage(screen, sprite.Texture.Img(), c, op)
		}

		// render icon for npcs
		if status.Role == model.NPC && global.Debug {
			geoM := status.RoleTexture().GetGeoM()
			geoM.Rotate(camera.Rotation)
			geoM.Translate(pos[0], pos[1])
			geoM.Concat(wordM)
			op := &colorm.DrawImageOptions{}
			op.GeoM = geoM
			colorm.DrawImage(screen, status.RoleTexture().Img(), c, op)
		}
	}

	for _, instance := range sprite.Instances {
		if !instance.IsActive(tick) && instance.GetCast() == nil {
			continue
		}
		renderObject(instance.Face, instance.Object)
	}
}
