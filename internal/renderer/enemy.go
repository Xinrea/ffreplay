package renderer

import (
	"fmt"
	"image/color"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/furex/v2"
)

func (r *Renderer) EnemyRender(ecs *ecs.ECS, screen *ebiten.Image) {
	if !entry.GetGlobal().Loaded.Load() {
		return
	}

	for e := range tag.Enemy.Iter(ecs.World) {
		r.renderEnemy(ecs, screen, e)
	}
}

func (r *Renderer) renderEnemy(ecs *ecs.ECS, screen *ebiten.Image, enemy *donburi.Entry) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	global := component.Global.Get(tag.Global.MustFirst(ecs.World))
	tick := entry.GetTick()

	status := component.Status.Get(enemy)

	status.Render(tick, camera, screen, global.ShowTargetRing, global.Debug || !global.ReplayMode)

	// print some extra debug info
	if global.Debug {
		for _, instance := range status.Instances {
			if !instance.IsActive(tick) && instance.GetCast() == nil {
				continue
			}

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
}
