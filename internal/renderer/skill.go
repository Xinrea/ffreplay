package renderer

import (
	"image/color"

	"github.com/Xinrea/ffreplay/internal/component"
	"github.com/Xinrea/ffreplay/internal/entry"
	"github.com/Xinrea/ffreplay/internal/tag"
	"github.com/Xinrea/ffreplay/util"
	"github.com/fogleman/ease"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

func (r *Renderer) SkillRender(ecs *ecs.ECS, screen *ebiten.Image) {
	camera := component.Camera.Get(tag.Camera.MustFirst(ecs.World))
	worldGeo := camera.WorldMatrixInverted()
	for e := range tag.Skill.Iter(ecs.World) {
		skill := component.Skill.Get(e)
		// draw cast progress
		// progress := float32(skill.Time.StartTime+skill.Time.CastTime-util.Time()) / float32(skill.Time.CastTime) * 100
		//   |--------cast-------|
		//   | ------| -display- |
		//   | ----- | --------- |
		//   s      s+d         s+c
		// draw display range
		d := util.TickToMS(entry.GetTick(ecs) - skill.Time.StartTick)
		if skill.Time.CastTime-skill.Time.DisplayTime <= d {
			delta := d - skill.Time.CastTime + skill.Time.DisplayTime
			partTime := skill.Time.DisplayTime / 4
			param := 1.0
			if delta < partTime {
				param = ease.InOutQuart(float64(delta) / float64(partTime))
			}
			if skill.Time.DisplayTime-delta < 200 {
				param = ease.InOutQuart(float64(skill.Time.DisplayTime-delta) / 200)
			}
			// draw skill range
			scale := color.RGBA{255, 255, 255, uint8(255 * param)}
			skill.GameSkill.Range().Render(screen, worldGeo, scale)
		}
	}
}
