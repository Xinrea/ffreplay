package renderer

import (
	"fmt"
	"image/color"
	"log"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/internal/ui"
	"github.com/Xinrea/ffreplay/pkg/texture"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/furex/v2"
)

const (
	TIMELINE_WIDTH  = 800
	TIMELINE_HEIGHT = 150
	NONEGCD_GAP     = 50
)

var (
	background     *ebiten.Image = nil
	skillLayer     *ebiten.Image = nil
	skillLayerMask *ebiten.Image = nil
)

type SkillTimeline struct {
	Periods []GCDPeriod
}

type GCDPeriod struct {
	StartTick     int64
	GCDSkill      *model.Skill
	NoneGCDSkills []*model.Skill
}

func initBackground() {
	s := ebiten.Monitor().DeviceScaleFactor()
	background = ebiten.NewImage(int(TIMELINE_WIDTH*s), int(TIMELINE_HEIGHT*s))
	skillLayer = ebiten.NewImage(int(TIMELINE_WIDTH*s), int(TIMELINE_HEIGHT*s))
	skillLayerMask = ebiten.NewImage(int(TIMELINE_WIDTH*s), int(TIMELINE_HEIGHT*s))

	for i := 0; i < int(TIMELINE_WIDTH*s); i++ {
		p := 1.0

		if i < int(TIMELINE_WIDTH*s*0.16) {
			p = float64(i) / (TIMELINE_WIDTH * s * 0.16)
		}

		if i > int(TIMELINE_WIDTH*s*0.84) {
			p = 1 - (float64(i)-TIMELINE_WIDTH*s*0.84)/(TIMELINE_WIDTH*s*0.16)
		}

		for j := 0; j < int(TIMELINE_HEIGHT*s); j++ {
			background.Set(i, j, color.NRGBA{0, 0, 0, uint8(p * 128)})
		}
	}

	for i := 0; i < int(TIMELINE_WIDTH*s); i++ {
		p := 1.0

		if i < int(TIMELINE_WIDTH*s*0.16) {
			p = float64(i) / (TIMELINE_WIDTH * s * 0.16)
		}

		if i > int(TIMELINE_WIDTH*s*0.84) {
			p = 1 - (float64(i)-TIMELINE_WIDTH*s*0.84)/(TIMELINE_WIDTH*s*0.16)
		}

		for j := 0; j < int(TIMELINE_HEIGHT*s); j++ {
			skillLayerMask.Set(i, j, color.NRGBA{0, 0, 0, uint8(p * 255)})
		}
	}
}

func RenderCasting(debug bool, canvas *ebiten.Image, tick int64, cast *model.Skill, x, y float64) {
	s := ebiten.Monitor().DeviceScaleFactor()
	textSize := 12.0 * s
	yOffset := 30.0 * s

	iconTexture := cast.Texture()
	geoM := texture.CenterGeoM(iconTexture)
	geoM.Scale(s, s)

	borderGeoM := texture.CenterGeoM(model.BorderTexture)
	borderGeoM.Scale(s, s)

	if !cast.IsGCD {
		geoM.Scale(0.8, 0.8)
		borderGeoM.Scale(0.8, 0.8)
		geoM.Translate(0, -30*s)
		borderGeoM.Translate(0, -30*s)
		textSize = 10.0 * s
		yOffset = -55 * s
	}

	geoM.Translate(x, y)
	borderGeoM.Translate(x, y)
	canvas.DrawImage(iconTexture, &ebiten.DrawImageOptions{GeoM: geoM})
	canvas.DrawImage(model.BorderTexture, &ebiten.DrawImageOptions{GeoM: borderGeoM})

	name := cast.Name
	if debug {
		name = fmt.Sprintf("[%d]%s", cast.ID, cast.Name)
	}

	ui.DrawText(canvas, name, textSize, x, y+yOffset, color.White, furex.AlignItemCenter, textShdowOpt)
}

func (g GCDPeriod) Render(debug bool, canvas *ebiten.Image, x, y float64, tick int64) {
	s := ebiten.Monitor().DeviceScaleFactor()

	if g.GCDSkill != nil {
		// tailLength := tickToLength(util.MSToTick(g.GCDSkill.Cast))
		// DrawFilledRect(canvas, x, y-6, tailLength, 12, color.NRGBA{119, 123, 131, 255})
		RenderCasting(debug, canvas, tick, g.GCDSkill, x, y)
	}

	for i := range g.NoneGCDSkills {
		RenderCasting(debug, canvas, tick, g.NoneGCDSkills[i], x+float64(i+1)*NONEGCD_GAP*s, y)
	}
}

func NewSkillTimeline(casts []*model.Skill) SkillTimeline {
	periods := []GCDPeriod{}
	currentPeriod := GCDPeriod{
		StartTick: -1,
	}

	for _, c := range casts {
		if c.IsGCD {
			if currentPeriod.StartTick != -1 {
				periods = append(periods, currentPeriod)
			}

			currentPeriod = GCDPeriod{
				StartTick: c.StartTick,
				GCDSkill:  c,
			}

			continue
		}

		if currentPeriod.StartTick == -1 {
			currentPeriod.StartTick = c.StartTick
		}

		currentPeriod.NoneGCDSkills = append(currentPeriod.NoneGCDSkills, c)
	}

	if currentPeriod.StartTick != -1 {
		periods = append(periods, currentPeriod)
	}

	return SkillTimeline{
		Periods: periods,
	}
}

func tickToLength(tick int64) float64 {
	return float64(tick) * 1.3
}

func (st SkillTimeline) Render(debug bool, canvas *ebiten.Image, x, y float64, tick int64) {
	s := ebiten.Monitor().DeviceScaleFactor()

	if len(st.Periods) == 0 {
		return
	}

	if int(TIMELINE_WIDTH*s) != background.Bounds().Dx() {
		initBackground()
		log.Println("Scale factor changed, recreated skill timeline assets")
	}

	x = x * s
	y = y * s

	x = x - TIMELINE_WIDTH*s/2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)

	canvas.DrawImage(background, op)

	skillLayer.Clear()

	for i := range st.Periods {
		offset := tickToLength(tick - st.Periods[i].StartTick)
		px := -offset*1.3*s + TIMELINE_WIDTH*s
		py := TIMELINE_HEIGHT*s/2 + 15.0*s
		st.Periods[i].Render(debug, skillLayer, px, py, tick)
	}

	skillLayer.DrawImage(skillLayerMask, &ebiten.DrawImageOptions{Blend: ebiten.BlendDestinationIn})
	canvas.DrawImage(skillLayer, op)
}
