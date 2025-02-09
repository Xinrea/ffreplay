package renderer

import (
	"fmt"
	"image/color"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model/role"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/pkg/object"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
)

func (r *Renderer) EnemyRender(ecs *ecs.ECS, screen *ebiten.Image) {
	if !entry.GetGlobal(ecs).Loaded.Load() {
		return
	}

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

		wordM := camera.WorldMatrixInverted()

		var c colorm.ColorM
		// render target ring
		if global.Debug || status.Role == role.Boss || global.RenderNPC {
			geoM := texture.CenterGeoM(sprite.Texture)
			if status.Role == role.NPC {
				geoM.Scale(0.5, 0.5)
			}

			geoM.Rotate(face)
			geoM.Translate(pos[0], pos[1])
			geoM.Concat(wordM)

			op := &colorm.DrawImageOptions{}
			op.GeoM = geoM

			colorm.DrawImage(screen, sprite.Texture, c, op)
		}
	}

	for _, instance := range sprite.Instances {
		if !instance.IsActive(tick) && instance.GetCast() == nil {
			continue
		}

		renderObject(instance.Face, instance.Object)

		if global.Debug && instance.GetCast() != nil {
			// render casting skill name
			cast := instance.GetCast()
			if cast != nil {
				px, py := camera.WorldToScreen(instance.Object.Position()[0], instance.Object.Position()[1])
				ui.DrawText(
					screen,
					fmt.Sprintf("[%d]%s", cast.ID, cast.Name),
					12,
					px,
					py,
					color.White,
					furex.AlignItemCenter,
					nil,
				)
			}
		}
	}
}
