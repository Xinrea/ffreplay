package renderer

import (
	"fmt"
	"image/color"
	"log"

	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/util"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	TIMELINE_WIDTH  = 800
	TIMELINE_HEIGHT = 150
	NONEGCD_GAP     = 50
)

var background *ebiten.Image = nil
var skillLayer *ebiten.Image = nil
var skillLayerMask *ebiten.Image = nil

type SkillTimeline struct {
	Periods []GCDPeriod
}

type GCDPeriod struct {
	StartTick     int64
	GCDSkill      *model.Skill
	NoneGCDSkills []*model.Skill
}

func init() {
	initBackground()
}

func initBackground() {
	s := ebiten.Monitor().DeviceScaleFactor()
	background = ebiten.NewImage(int(TIMELINE_WIDTH*s), int(TIMELINE_HEIGHT*s))
	skillLayer = ebiten.NewImage(int(TIMELINE_WIDTH*s), int(TIMELINE_HEIGHT*s))
	skillLayerMask = ebiten.NewImage(int(TIMELINE_WIDTH*s), int(TIMELINE_HEIGHT*s))
	for i := 0; i < int(TIMELINE_WIDTH*s); i++ {
		p := min(128, min(i, int(TIMELINE_WIDTH*s)-i))
		for j := 0; j < int(TIMELINE_HEIGHT*s); j++ {
			background.Set(i, j, color.NRGBA{0, 0, 0, uint8(p)})
		}
	}
	for i := 0; i < int(TIMELINE_WIDTH*s); i++ {
		p := min(255, min(i, int(TIMELINE_WIDTH*s)-i))
		for j := 0; j < int(TIMELINE_HEIGHT*s); j++ {
			skillLayerMask.Set(i, j, color.NRGBA{0, 0, 0, uint8(p)})
		}
	}
}

func RenderCasting(debug bool, canvas *ebiten.Image, tick int64, cast *model.Skill, x, y float64) {
	s := ebiten.Monitor().DeviceScaleFactor()
	textSize := 12.0
	yOffset := 20.0
	iconTexture := cast.Texture()
	geoM := iconTexture.GetGeoM()
	borderGeoM := model.BorderGeoM
	if !model.IsGCD(cast.ID) {
		geoM.Scale(0.8, 0.8)
		borderGeoM.Scale(0.8, 0.8)
		geoM.Translate(0, -30)
		borderGeoM.Translate(0, -30)
		textSize = 10.0
		yOffset = -60
	}
	geoM.Translate(x, y)
	geoM.Scale(s, s)
	borderGeoM.Translate(x, y)
	borderGeoM.Scale(s, s)
	canvas.DrawImage(iconTexture.Img(), &ebiten.DrawImageOptions{GeoM: geoM})
	canvas.DrawImage(model.BorderTexture.Img(), &ebiten.DrawImageOptions{GeoM: borderGeoM})
	name := cast.Name
	if debug {
		name = fmt.Sprintf("[%d]%s", cast.ID, cast.Name)
	}
	DrawText(canvas, name, textSize, x, y+yOffset, color.White, AlignCenter)
}

func (g GCDPeriod) Render(debug bool, canvas *ebiten.Image, x, y float64, tick int64) {
	if g.GCDSkill != nil {
		tailLength := tickToLength(util.MSToTick(g.GCDSkill.Cast))
		DrawFilledRect(canvas, x, y-6, tailLength, 12, color.NRGBA{119, 123, 131, 255})
		RenderCasting(debug, canvas, tick, g.GCDSkill, x, y)
	}
	for i := range g.NoneGCDSkills {
		RenderCasting(debug, canvas, tick, g.NoneGCDSkills[i], x+float64(i+1)*NONEGCD_GAP, y)
	}
}

func NewSkillTimeline(casts []*model.Skill) SkillTimeline {
	periods := []GCDPeriod{}
	currentPeriod := GCDPeriod{
		StartTick: -1,
	}
	for _, c := range casts {
		if model.IsGCD(c.ID) {
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
	if len(st.Periods) == 0 {
		return
	}
	s := ebiten.Monitor().DeviceScaleFactor()
	if int(s*TIMELINE_WIDTH) != background.Bounds().Dx() {
		initBackground()
		log.Println("Scale factor changed, recreated skill timeline assets")
	}
	x = x - TIMELINE_WIDTH/2
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x*s, y*s)
	canvas.DrawImage(background, op)

	skillLayer.Clear()
	for i := range st.Periods {
		offset := tickToLength(tick - st.Periods[i].StartTick)
		// draw seperator
		px := -offset*1.3 + TIMELINE_WIDTH
		py := TIMELINE_HEIGHT/2 + 20.0
		st.Periods[i].Render(debug, skillLayer, px, py, tick)
	}
	skillLayer.DrawImage(skillLayerMask, &ebiten.DrawImageOptions{Blend: ebiten.BlendDestinationIn})
	canvas.DrawImage(skillLayer, op)
}
