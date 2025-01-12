package renderer

import (
	"image/color"
	"log"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/fogleman/ease"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) RangeRender(ecs *ecs.ECS, screen *ebiten.Image) {
	tick := entry.GetTick(ecs)

	for e := range tag.Timeline.Iter(ecs.World) {
		timeline := component.Timeline.Get(e)
		if timeline.IsDone(tick) {
			continue
		}

		timelineRender(ecs, screen, timeline, tick)
	}
}

func timelineRender(ecs *ecs.ECS, screen *ebiten.Image, timeline *model.TimelineData, tick int64) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	worldGeo := camera.WorldMatrixInverted()

	for i := range timeline.Events {
		if timeline.Events[i].EffectRange == nil {
			continue
		}

		current := tick - timeline.StartTick - timeline.Events[i].OffsetTick()
		if current > timeline.Events[i].DisplayTick() || current < 0 {
			continue
		}

		param := 1.0

		partTick := timeline.Events[i].DisplayTick() / 4
		if current < partTick {
			param = ease.InOutQuart(float64(current) / float64(partTick))
		}

		if timeline.Events[i].DisplayTick()-current <= partTick {
			param = ease.InOutQuart(float64(timeline.Events[i].DisplayTick()-current) / float64(partTick))
		}

		if param < 0 || param > 1 {
			log.Panic("Invalid param", current, partTick, param)
		}
		// draw skill range
		scale := color.RGBA{255, 255, 255, uint8(255 * param)}
		timeline.Events[i].EffectRange.Render(screen, worldGeo, scale)
	}
}
