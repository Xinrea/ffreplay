package renderer

import (
	"image/color"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/fogleman/ease"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) RangeRender(ecs *ecs.ECS, screen *ebiten.Image) {
	if !entry.GetGlobal().RangeDisplay {
		return
	}

	tick := entry.GetTick()

	for entry := range component.Status.Iter(ecs.World) {
		status := component.Status.Get(entry)

		for _, inst := range status.Instances {
			skill := inst.GetCast()
			if skill == nil {
				continue
			}

			if skill.EffectRange == nil {
				continue
			}

			rangeRender(ecs, screen, skill, tick)
		}
	}
}

func rangeRender(ecs *ecs.ECS, screen *ebiten.Image, skill *model.Skill, tick int64) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	worldGeo := camera.WorldMatrixInverted()

	current := tick - skill.StartTick
	castTick := util.MSToTick(skill.Cast)

	param := 1.0

	partTick := min(castTick/4, 30)
	if current < partTick {
		param = ease.InOutQuart(float64(current) / float64(partTick))
	}

	if castTick-current <= partTick {
		param = ease.InOutQuart(float64(castTick-current) / float64(partTick))
	}

	param = util.Clamp(param, 0, 1)

	// draw skill range
	scale := color.RGBA{255, 255, 255, uint8(255 * param)}
	skill.EffectRange.Render(screen, worldGeo, scale)
}
